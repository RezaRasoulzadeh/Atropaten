package domain

import (
	"fmt"
	"sort"
	"strings"
)

// MaterialOption is a derived option. Its value is either a canonical
// attribute value or an explicit material ID; it is never used to infer an ID
// from a display label.
type MaterialOption struct {
	Value       string
	Label       string
	MaterialIDs []string
}

// CompatibleMaterials applies every explicit material-backed parameter as an
// intersection. Parameter ordering is intentionally irrelevant.
func CompatibleMaterials(service Service, materials []Material, selected map[string]string) ([]Material, error) {
	return compatibleMaterials(service, materials, selected, "")
}

func compatibleMaterials(service Service, materials []Material, selected map[string]string, skipKey string) ([]Material, error) {
	if err := service.Validate(); err != nil {
		return nil, err
	}
	result := make([]Material, 0, len(materials))
	for _, material := range materials {
		if !material.Active {
			continue
		}
		compatible := true
		for _, parameter := range service.Parameters {
			if parameter.Type == ParameterMaterialReference && parameter.Key != skipKey {
				value := strings.TrimSpace(selected[parameter.Key])
				if parameter.Key == skipKey {
					value = ""
				}
				if value != "" && value != material.ID {
					compatible = false
					break
				}
			}
			source := parameter.MaterialSource
			if source == nil {
				continue
			}
			if len(source.AllowedKinds) > 0 && !containsKind(source.AllowedKinds, material.Kind) {
				compatible = false
				break
			}
			if !matchesFilters(material, source.AdditionalFilters) {
				compatible = false
				break
			}
			if len(source.AttributeKeys()) > 0 {
				if !sourceValueAllowed(material, *source) {
					compatible = false
					break
				}
				value := strings.TrimSpace(selected[parameter.Key])
				if parameter.Key == skipKey {
					value = ""
				}
				if value != "" && !source.SelectMaterial && !sourceValueMatches(material, *source, value) {
					compatible = false
					break
				}
			}
			if source.SelectMaterial {
				value := strings.TrimSpace(selected[parameter.Key])
				if parameter.Key == skipKey {
					value = ""
				}
				if value != "" && value != material.ID {
					compatible = false
					break
				}
			}
		}
		if compatible && !finishedDimensionsFit(service, material, selected) {
			compatible = false
		}
		if compatible {
			result = append(result, material)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, right := strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)
		if left == right {
			return result[i].ID < result[j].ID
		}
		return left < right
	})
	return result, nil
}

func finishedDimensionsFit(service Service, material Material, selected map[string]string) bool {
	if service.FinishedSize == nil {
		return true
	}
	width, height, err := service.ResolveFinishedDimensions(selected)
	if err != nil {
		// An empty size is normal while an order is being configured. Once the
		// operator has supplied a size value, malformed or unavailable
		// dimensions must make the material unavailable instead of silently
		// passing through to pricing.
		if service.FinishedSize.AllowCustom {
			return strings.TrimSpace(selected[service.FinishedSize.WidthParameterKey]) == "" && strings.TrimSpace(selected[service.FinishedSize.HeightParameterKey]) == ""
		}
		return strings.TrimSpace(selected[service.FinishedSize.ParameterKey]) == ""
	}
	switch material.Kind {
	case MaterialKindSheetStock, MaterialKindBoard:
		sheetWidth, sheetHeight, ok := material.PhysicalDimensions()
		if !ok {
			return false
		}
		if sheetWidth/width > 0 && sheetHeight/height > 0 {
			return true
		}
		return service.FinishedSize.AllowRotation && sheetWidth/height > 0 && sheetHeight/width > 0
	case MaterialKindRollMedia, MaterialKindFabric:
		rollWidth, ok := material.RollWidthMM()
		if !ok {
			return false
		}
		return rollWidth/width > 0 || (service.FinishedSize.AllowRotation && rollWidth/height > 0)
	default:
		return true
	}
}

// ResolveMaterialSelection returns the unique material left by the service
// constraints. A direct material selection may make the result unique; an
// ambiguous configuration is rejected rather than guessed from labels.
func ResolveMaterialSelection(service Service, materials []Material, selected map[string]string) (Material, error) {
	if len(service.MaterialVariants) > 0 {
		variant, ok := service.ResolveMaterialVariant(selected)
		if !ok {
			return Material{}, fmt.Errorf("selected material combination is unavailable")
		}
		for _, material := range materials {
			if material.Active && material.ID == variant.MaterialID {
				return material, nil
			}
		}
		return Material{}, fmt.Errorf("configured material %q is unavailable", variant.MaterialID)
	}
	compatible, err := CompatibleMaterials(service, materials, selected)
	if err != nil {
		return Material{}, err
	}
	if len(compatible) == 0 {
		return Material{}, fmt.Errorf("no compatible active material")
	}
	if len(compatible) > 1 {
		return Material{}, fmt.Errorf("material selection is ambiguous: %d compatible materials remain", len(compatible))
	}
	return compatible[0], nil
}

// ResolveMaterialVariant returns the exact service mapping for a complete
// option selection. Matching is by stable parameter values and never by
// display labels. A variant may contain a subset of parameters so services
// can model optional groups, but every supplied value must match exactly.
func (s Service) ResolveMaterialVariant(selected map[string]string) (ServiceMaterialVariant, bool) {
	for _, variant := range s.MaterialVariants {
		if !variant.Active {
			continue
		}
		matches := true
		for key, value := range variant.Values {
			if strings.TrimSpace(selected[key]) != strings.TrimSpace(value) {
				matches = false
				break
			}
		}
		if matches {
			return variant, true
		}
	}
	return ServiceMaterialVariant{}, false
}

// MaterialOptionsForParameter derives values after applying all other
// material-backed selections. This keeps dependent options honest.
func MaterialOptionsForParameter(service Service, parameterKey string, materials []Material, selected map[string]string) ([]MaterialOption, error) {
	parameterIndex := -1
	for i, parameter := range service.Parameters {
		if parameter.Key == parameterKey {
			parameterIndex = i
			break
		}
	}
	if parameterIndex < 0 {
		return nil, fmt.Errorf("material parameter %q not found", parameterKey)
	}
	parameter := service.Parameters[parameterIndex]
	if parameter.MaterialSource == nil {
		return nil, fmt.Errorf("parameter %q is not material-backed", parameterKey)
	}
	// Once a service has explicit mappings, derive the option list from those
	// mappings rather than from every material that happens to share an
	// attribute. This keeps the configurator and pricing resolver aligned: an
	// option cannot be displayed unless the operator configured a material for
	// the complete combination.
	if len(service.MaterialVariants) > 0 {
		options := make([]MaterialOption, 0)
		seen := map[string]int{}
		for _, variant := range service.MaterialVariants {
			if !variant.Active {
				continue
			}
			value, ok := variant.Values[parameterKey]
			if !ok || strings.TrimSpace(value) == "" {
				continue
			}
			matches := true
			for key, selectedValue := range selected {
				selectedValue = strings.TrimSpace(selectedValue)
				if key == parameterKey || selectedValue == "" {
					continue
				}
				variantValue, exists := variant.Values[key]
				if !exists || strings.TrimSpace(variantValue) != selectedValue {
					matches = false
					break
				}
			}
			if !matches {
				continue
			}
			var material Material
			found := false
			for _, candidate := range materials {
				if candidate.Active && candidate.ID == variant.MaterialID {
					material, found = candidate, true
					break
				}
			}
			if !found {
				continue
			}
			merged := make(map[string]string, len(selected)+1)
			for key, selectedValue := range selected {
				merged[key] = selectedValue
			}
			merged[parameterKey] = strings.TrimSpace(value)
			if compatible, err := CompatibleMaterials(service, []Material{material}, merged); err != nil || len(compatible) == 0 {
				continue
			}
			label := strings.TrimSpace(value)
			if parameter.MaterialSource.SelectMaterial {
				label = material.Name
			}
			if index, exists := seen[value]; exists {
				options[index].MaterialIDs = append(options[index].MaterialIDs, material.ID)
				continue
			}
			seen[value] = len(options)
			options = append(options, MaterialOption{Value: strings.TrimSpace(value), Label: label, MaterialIDs: []string{material.ID}})
		}
		sort.SliceStable(options, func(i, j int) bool { return options[i].Value < options[j].Value })
		return options, nil
	}
	withoutCurrent := make(map[string]string, len(selected))
	for key, value := range selected {
		if key != parameterKey {
			withoutCurrent[key] = value
		}
	}
	compatible, err := CompatibleMaterialsWithSkippedParameter(service, parameterKey, materials, withoutCurrent)
	if err != nil {
		return nil, err
	}
	options := make([]MaterialOption, 0)
	seen := map[string]int{}
	for _, material := range compatible {
		value := material.ID
		label := material.Name
		if len(parameter.MaterialSource.AttributeKeys()) > 0 && !parameter.MaterialSource.SelectMaterial {
			derived, ok := sourceValue(material, *parameter.MaterialSource)
			if !ok {
				continue
			}
			value = derived
			label = value
		}
		if index, ok := seen[value]; ok {
			options[index].MaterialIDs = append(options[index].MaterialIDs, material.ID)
			continue
		}
		seen[value] = len(options)
		options = append(options, MaterialOption{Value: value, Label: label, MaterialIDs: []string{material.ID}})
	}
	sort.SliceStable(options, func(i, j int) bool { return options[i].Value < options[j].Value })
	return options, nil
}

func CompatibleMaterialsWithSkippedParameter(service Service, skipKey string, materials []Material, selected map[string]string) ([]Material, error) {
	return compatibleMaterials(service, materials, selected, skipKey)
}

func materialAttribute(material Material, key string) (MaterialAttributeValue, bool) {
	for _, attribute := range material.Attributes {
		if attribute.Key == key {
			return attribute, true
		}
	}
	return MaterialAttributeValue{}, false
}

func sourceValue(material Material, source MaterialParameterSource) (string, bool) {
	keys := source.AttributeKeys()
	if len(keys) == 0 {
		return "", false
	}
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		attribute, ok := materialAttribute(material, key)
		if !ok {
			return "", false
		}
		values = append(values, attribute.Canonical())
	}
	return strings.Join(values, "\x1f"), true
}

func sourceValueMatches(material Material, source MaterialParameterSource, selected string) bool {
	value, ok := sourceValue(material, source)
	return ok && strings.TrimSpace(value) == strings.TrimSpace(selected)
}

func sourceValueAllowed(material Material, source MaterialParameterSource) bool {
	keys := source.AttributeKeys()
	if len(keys) == 0 {
		return true
	}
	for _, key := range keys {
		if _, ok := materialAttribute(material, key); !ok {
			return false
		}
	}
	if len(source.AllowedValues) == 0 {
		return true
	}
	for _, key := range keys {
		matchedKey := false
		for _, allowed := range source.AllowedValues {
			if allowed.Key != key {
				continue
			}
			attribute, ok := materialAttribute(material, key)
			if ok && attributeValuesEqual(attribute, allowed) {
				matchedKey = true
				break
			}
		}
		if !matchedKey {
			for _, allowed := range source.AllowedValues {
				if allowed.Key == key {
					return false
				}
			}
		}
	}
	return true
}

func matchesFilters(material Material, filters []MaterialAttributeFilter) bool {
	for _, filter := range filters {
		attribute, ok := materialAttribute(material, filter.Key)
		if !ok || !attributeValuesEqual(attribute, filter.Value) {
			return false
		}
	}
	return true
}

func containsAttributeValue(values []MaterialAttributeValue, wanted MaterialAttributeValue) bool {
	for _, value := range values {
		if attributeValuesEqual(value, wanted) {
			return true
		}
	}
	return false
}

func attributeValuesEqual(left, right MaterialAttributeValue) bool {
	return left.ValueType == right.ValueType && left.Canonical() == right.Canonical()
}

func attributeMatchesString(attribute MaterialAttributeValue, value string) bool {
	return attribute.Canonical() == strings.TrimSpace(value)
}

func containsKind(kinds []MaterialKind, wanted MaterialKind) bool {
	for _, kind := range kinds {
		if kind == wanted {
			return true
		}
	}
	return false
}
