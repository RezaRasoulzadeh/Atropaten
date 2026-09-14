package domain

import "strings"

// PredefinedParameterPrintSize is the first shared parameter definition. Its
// option codes are stable data keys; labels and dimensions are catalog data.
const PredefinedParameterPrintSize = "print_size"
const PredefinedParameterPaperType = "paper_type"
const PredefinedParameterColor = "color"

func IsSupportedPredefinedParameter(key string) bool {
	return key == PredefinedParameterPrintSize || key == PredefinedParameterPaperType || key == PredefinedParameterColor
}

type PredefinedParameterOption struct {
	Code     string
	Label    string
	WidthMM  *Quantity
	HeightMM *Quantity
	Active   bool
	Position int
}

type PredefinedParameterDefinition struct {
	Key       string
	Label     string
	ValueType ParameterType
	Unit      string
	Active    bool
	Position  int
	Options   []PredefinedParameterOption
}

func (d PredefinedParameterDefinition) Validate() error {
	if strings.TrimSpace(d.Key) == "" || strings.TrimSpace(d.Label) == "" {
		return validationError("predefinedParameter", "requires a key and label")
	}
	if d.ValueType != ParameterChoice {
		return validationError("predefinedParameter.valueType", "must be a choice")
	}
	seen := map[string]struct{}{}
	for index, option := range d.Options {
		if strings.TrimSpace(option.Code) == "" || strings.TrimSpace(option.Label) == "" {
			return validationError("predefinedParameter.options", "requires code and label")
		}
		if option.Position != index {
			return validationError("predefinedParameter.options.position", "must be deterministic")
		}
		if _, exists := seen[option.Code]; exists {
			return validationError("predefinedParameter.options", "codes must be unique")
		}
		seen[option.Code] = struct{}{}
		if (option.WidthMM == nil) != (option.HeightMM == nil) {
			return validationError("predefinedParameter.options", "dimensions must include both width and height")
		}
		if option.WidthMM != nil && (*option.WidthMM <= 0 || *option.HeightMM <= 0) {
			return validationError("predefinedParameter.options", "dimensions must be positive")
		}
	}
	return nil
}

func PrintSizePredefinedParameter(options []PredefinedParameterOption) PredefinedParameterDefinition {
	return PredefinedParameterDefinition{Key: PredefinedParameterPrintSize, Label: "Print size", ValueType: ParameterChoice, Unit: "mm", Active: true, Options: options}
}

func PaperTypePredefinedParameter(options []PredefinedParameterOption) PredefinedParameterDefinition {
	return PredefinedParameterDefinition{Key: PredefinedParameterPaperType, Label: "Paper type", ValueType: ParameterChoice, Active: true, Options: options}
}

func ColorPredefinedParameter(options []PredefinedParameterOption) PredefinedParameterDefinition {
	return PredefinedParameterDefinition{Key: PredefinedParameterColor, Label: "Color", ValueType: ParameterChoice, Active: true, Options: options}
}
