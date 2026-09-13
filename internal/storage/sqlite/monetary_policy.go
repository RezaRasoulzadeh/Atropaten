package sqlite

import (
	"context"

	"Atropaten/internal/domain"
)

// roundCalculatedMoney applies the current policy to a newly calculated
// projection. It never writes the rounded value back to a historical record.
func (s *Store) roundCalculatedMoney(ctx context.Context, amount int64) (int64, error) {
	settings, err := s.GetShopSettings(ctx)
	if err != nil {
		return 0, err
	}
	step := settings.MonetaryRoundingStepRial
	if step <= 0 {
		step = domain.DefaultMonetaryRoundingStepRial
	}
	return domain.RoundMoneyUp(amount, step)
}
