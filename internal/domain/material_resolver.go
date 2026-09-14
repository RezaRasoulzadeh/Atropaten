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
			if source.ExposedAttributeKey != "" {
				attribute, ok := materialAttribute(material, source.ExposedAttributeKey)
				if !ok {
					compatible = false
					break
				}
				if len(source.AllowedValues) > 0 && !containsAttributeValue(source.AllowedValues, attribute) {
					compatible = false
					break
				}
				value := strings.TrimSpace(selected[parameter.Key])
				if parameter.Key == skipKey {
					value = ""
				}
				if value != "" && !source.SelectMaterial && !attributeMatchesString(attribute, value) {
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
		return true
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
		if parameter.MaterialSource.ExposedAttributeKey != "" && !parameter.MaterialSource.SelectMaterial {
			attribute, ok := materialAttribute(material, parameter.MaterialSource.ExposedAttributeKey)
			if !ok {
				continue
			}
			value, label = attribute.Canonical(), attribute.Canonical()
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
