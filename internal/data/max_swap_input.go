package data

import (
	"math/big"
)

func (c *Chain) MaxSwapInput(
	userInBalance *big.Int,
	inTokenLimit *big.Int,
	poolOutBalance *big.Int,
	inRate uint64,
	outRate uint64,
	inDecimals uint8,
	outDecimals uint8,
) *big.Int {
	if poolOutBalance.Sign() == 0 {
		return big.NewInt(0)
	}

	numerator := new(big.Int).Set(poolOutBalance)
	numerator.Mul(numerator, new(big.Int).SetUint64(outRate))
	inMultiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(inDecimals)), nil)
	numerator.Mul(numerator, inMultiplier)

	outMultiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(outDecimals)), nil)
	denominator := new(big.Int).Mul(new(big.Int).SetUint64(inRate), outMultiplier)
	maxFromOut := new(big.Int).Div(numerator, denominator)

	// Check in balance
	maxInput := new(big.Int).Set(userInBalance)

	// Check limit
	if inTokenLimit.Cmp(maxInput) < 0 {
		maxInput.Set(inTokenLimit)
	}

	// Check pool out balance
	if maxFromOut.Cmp(maxInput) < 0 {
		maxInput.Set(maxFromOut)
	}

	return maxInput
}
