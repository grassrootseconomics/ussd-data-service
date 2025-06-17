package data

import (
	"fmt"
	"math/big"
	"testing"
)

func bigFromString(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(fmt.Sprintf("failed to parse string to big.Int: %s", s))
	}
	return n
}

var chainClient = &Chain{}

func TestMaxSwapInput(t *testing.T) {
	tokaDecimals := uint8(18)
	tokbDecimals := uint8(6)
	tokaRate := uint64(129000)
	tokbRate := uint64(10000)

	testCases := []struct {
		name           string
		userInBalance  *big.Int
		inTokenLimit   *big.Int
		poolOutBalance *big.Int
		inRate         uint64
		outRate        uint64
		inDecimals     uint8
		outDecimals    uint8
		expected       *big.Int
	}{
		{
			name:           "0.64 cUSD -> RIBA, Max is user balance",
			userInBalance:  bigFromString("640000000000000000"),
			inTokenLimit:   bigFromString("10000000000000000000000"),
			poolOutBalance: bigFromString("271299506"),
			inRate:         tokaRate,
			outRate:        tokbRate,
			inDecimals:     tokaDecimals,
			outDecimals:    tokbDecimals,
			expected:       bigFromString("640000000000000000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := chainClient.MaxSwapInput(
				tc.userInBalance,
				tc.inTokenLimit,
				tc.poolOutBalance,
				tc.inRate,
				tc.outRate,
				tc.inDecimals,
				tc.outDecimals,
			)

			if got.Cmp(tc.expected) != 0 {
				t.Errorf("MaxSwapInput() failed\nExpected: %s\nGot:      %s", tc.expected.String(), got.String())
			}
		})
	}
}
