package data

import (
	"math/big"
)

func (c *Chain) MaxSwapInput(
	userInBalance *big.Int,
	inTokenLimit *big.Int,
	outTokenLimit *big.Int,
	poolInBalance *big.Int,
	poolOutBalance *big.Int,
	inRate uint64,
	outRate uint64,
	inDecimals uint8,
	outDecimals uint8,
) *big.Int {
	holdingA := new(big.Int).Sub(inTokenLimit, poolInBalance)
	if holdingA.Sign() < 0 {
		holdingA = big.NewInt(0)
	}

	maxOutputB := new(big.Int).Set(poolOutBalance)

	if maxOutputB.Sign() == 0 {
		return big.NewInt(0)
	}

	bigInRate := new(big.Int).SetUint64(inRate)
	bigOutRate := new(big.Int).SetUint64(outRate)

	pow10In := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(inDecimals)), nil)
	pow10Out := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(outDecimals)), nil)

	numerator := new(big.Int).Mul(maxOutputB, bigOutRate)
	numerator.Mul(numerator, pow10In)

	denominator := new(big.Int).Mul(bigInRate, pow10Out)

	if denominator.Sign() == 0 {
		return big.NewInt(0)
	}

	inputA_bound := new(big.Int).Div(numerator, denominator)

	maxInput := new(big.Int).Set(holdingA)
	if inputA_bound.Cmp(maxInput) < 0 {
		maxInput.Set(inputA_bound)
	}
	if userInBalance.Cmp(maxInput) < 0 {
		maxInput.Set(userInBalance)
	}

	return maxInput
}
