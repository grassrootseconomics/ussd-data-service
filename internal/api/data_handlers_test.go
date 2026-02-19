package api

import (
	"math/big"
	"testing"
)

func TestCalculateReverseQuote(t *testing.T) {
	tests := []struct {
		name         string
		outputAmount *big.Int
		inRate       uint64
		outRate      uint64
		inDecimals   uint8
		outDecimals  uint8
		want         *big.Int
	}{
		{
			name:         "real swap rates inRate=1290000 outRate=10000",
			outputAmount: big.NewInt(1010000),
			inRate:       1_290_000,
			outRate:      10_000,
			inDecimals:   6,
			outDecimals:  6,
			want:         big.NewInt(7829),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateReverseQuote(tt.outputAmount, tt.inRate, tt.outRate, tt.inDecimals, tt.outDecimals)
			t.Logf("outputAmount=%s, inRate=%d, outRate=%d, inDecimals=%d, outDecimals=%d => inputAmount=%s",
				tt.outputAmount.String(), tt.inRate, tt.outRate, tt.inDecimals, tt.outDecimals, got.String())
			if got.Cmp(tt.want) != 0 {
				t.Errorf("CalculateReverseQuote() = %s, want %s", got.String(), tt.want.String())
			}
		})
	}
}
