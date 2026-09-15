package domain

import "strings"

// WithRollSizeDefaults upgrades legacy roll definitions in memory. Reads and
// pricing must agree even before an older service has been edited and saved.
func (s Service) WithRollSizeDefaults() Service {
	roll := false
	switch strings.ToLower(strings.TrimSpace(s.Category)) {
	case "large format", "banners & signage":
		roll = true
	}
	for _, p := range s.Parameters {
		if !p.Active || p.MaterialSource == nil || len(p.MaterialSource.AllowedKinds) == 0 {
			continue
		}
		onlyRoll := true
		for _, kind := range p.MaterialSource.AllowedKinds {
			if kind != MaterialKindRollMedia && kind != MaterialKindFabric {
				onlyRoll = false
			}
		}
		roll = roll || onlyRoll
	}
	if !roll || (s.FinishedSize != nil && s.FinishedSize.QuantityParameterKey != "") {
		return s
	}
	size := ServiceFinishedSizeDefinition{WidthParameterKey: "finished_width_mm", HeightParameterKey: "finished_height_mm", AllowCustom: true, AllowRotation: true}
	if s.FinishedSize != nil && s.FinishedSize.AllowCustom {
		size = *s.FinishedSize
	}
	size.QuantityParameterKey = "layout_quantity"
	s.FinishedSize = &size
	s.Parameters = append([]ServiceParameter(nil), s.Parameters...)
	for _, spec := range []struct{ key, label, unit, value, min string }{
		{size.WidthParameterKey, "Finished width (mm)", "mm", "", "0.001"},
		{size.HeightParameterKey, "Custom height / length (mm)", "mm", "1000", "0.001"},
		{"layout_quantity", "Finished pieces", "piece", "1", "1"},
		{"layout_gap_mm", "Space between pieces (mm)", "mm", "0", "0"},
		{"layout_margin_mm", "Edge margin (mm)", "mm", "0", "0"},
	} {
		found := false
		for i := range s.Parameters {
			if s.Parameters[i].Key == spec.key {
				s.Parameters[i].Required = true
				found = true
				break
			}
		}
		if found {
			continue
		}
		kind := ParameterDecimal
		if spec.key == "layout_quantity" {
			kind = ParameterInteger
		}
		min, _ := ParseQuantity(spec.min)
		s.Parameters = append(s.Parameters, ServiceParameter{
			ID: s.ID + "-" + spec.key, ServiceID: s.ID, Key: spec.key, Label: spec.label,
			Type: kind, Required: true, Active: true, Position: len(s.Parameters),
			Unit: spec.unit, DefaultValue: spec.value, MinValue: &min,
			CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
		})
	}
	return s
}
