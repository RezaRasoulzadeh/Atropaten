package domain

import (
	"fmt"
	"math/big"
)

// DefaultMonetaryRoundingStepRial is the legacy customer-price increment
// (100 Toman = 1,000 Rial). It is seeded into settings so existing shops keep
// their current calculated selling-price behaviour after upgrading.
const DefaultMonetaryRoundingStepRial int64 = 1000

// RoundMoneyUp rounds a calculated monetary amount toward the next multiple
// of stepRial. It uses mathematical ceiling semantics for negative values:
// -12,001 with a 1,000-Rial step becomes -12,000. Big integers keep both the
// division and the final multiplication overflow-safe before returning Rial.
func RoundMoneyUp(amountRial, stepRial int64) (int64, error) {
	if stepRial <= 0 {
		return 0, fmt.Errorf("monetary rounding step must be positive")
	}
	if amountRial == 0 {
		return 0, nil
	}
	amount := big.NewInt(amountRial)
	step := big.NewInt(stepRial)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(amount, step, remainder)
	if amountRial > 0 && remainder.Sign() != 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	result := new(big.Int).Mul(quotient, step)
	if !result.IsInt64() {
		return 0, fmt.Errorf("rounded monetary result exceeds Rial range")
	}
	return result.Int64(), nil
}
