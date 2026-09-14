package domain

import (
	"fmt"
	"strings"
)

// ResolveFinishedDimensions resolves customer-facing production dimensions.
// These dimensions are deliberately independent from a material's physical
// inventory dimensions.
func (s Service) ResolveFinishedDimensions(selected map[string]string) (Quantity, Quantity, error) {
	if s.FinishedSize == nil {
		for _, parameter := range s.Parameters {
			if parameter.PredefinedKey != PredefinedParameterPrintSize {
				continue
			}
			code := strings.TrimSpace(selected[parameter.Key])
			var match *PredefinedParameterOption
			for index := range parameter.PredefinedOptions {
				option := &parameter.PredefinedOptions[index]
				if !option.Active || (code != "" && option.Code != code) {
					continue
				}
				if match != nil {
					return 0, 0, fmt.Errorf("print size selection is required")
				}
				match = option
			}
			if match == nil || match.WidthMM == nil || match.HeightMM == nil {
				return 0, 0, fmt.Errorf("print size %q has no finished dimensions", code)
			}
			return *match.WidthMM, *match.HeightMM, nil
		}
		return 0, 0, fmt.Errorf("service does not define finished dimensions")
	}
	definition := s.FinishedSize
	if definition.AllowCustom {
		width, err := ParseQuantity(strings.TrimSpace(selected[definition.WidthParameterKey]))
		if err != nil || width <= 0 {
			return 0, 0, fmt.Errorf("finished width must be positive")
		}
		height, err := ParseQuantity(strings.TrimSpace(selected[definition.HeightParameterKey]))
		if err != nil || height <= 0 {
			return 0, 0, fmt.Errorf("finished height must be positive")
		}
		return width, height, nil
	}
	code := strings.TrimSpace(selected[definition.ParameterKey])
	var match *FinishedSizeOption
	for index := range definition.Options {
		option := &definition.Options[index]
		if !option.Active {
			continue
		}
		if code == "" || option.Code == code {
			if match != nil {
				return 0, 0, fmt.Errorf("finished size selection is required")
			}
			match = option
		}
	}
	if match == nil {
		return 0, 0, fmt.Errorf("finished size selection %q is not available", code)
	}
	return match.WidthMM, match.HeightMM, nil
}
