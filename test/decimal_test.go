package test

import (
	"testing"

	"github.com/farseer-go/data/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewFromX(t *testing.T) {
	assert.Equal(t, decimal.NewFromString("13.4560").String(), "13.456")
	assert.Equal(t, decimal.NewFromString("100.000").String(), "100")
	assert.Equal(t, decimal.NewFromFloat64(13.456).Float64(), float64(13.456))
	assert.Equal(t, decimal.NewFromInt(13).IntPart(), int64(13))
	assert.Equal(t, decimal.NewFromInt32(13).IntPart(), int64(13))
	assert.Equal(t, decimal.NewFromInt64(13).IntPart(), int64(13))
}

func TestBasicOpr(t *testing.T) {
	assert.Equal(t, decimal.NewFromInt64(13).Add(decimal.NewFromInt64(5)).IntPart(), int64(18))
	assert.Equal(t, decimal.NewFromFloat64(13.88).Sub(decimal.NewFromFloat64(5.22)).Float64(), float64(8.66))
	assert.Equal(t, decimal.NewFromFloat64(13.88).Sub(decimal.NewFromFloat64(5.22)).Truncate().Float64(), float64(8))
	assert.Equal(t, decimal.NewFromFloat64(13.88).Sub(decimal.NewFromFloat64(5.22)).IntPart(), int64(8))
	assert.Equal(t, decimal.NewFromFloat64(13.88).MulFloat001().String(), "0.1388")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Mul(decimal.NewFromFloat64(2)).String(), "27.76")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Div(decimal.NewFromFloat64(2)).String(), "6.94")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Abs().StringFixed(0), "14")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Abs().StringFixed(1), "13.9")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Abs().StringFixed(2), "13.88")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Abs().StringFixed(3), "13.880")
	assert.False(t, decimal.NewFromFloat64(13.88).Greater(decimal.NewFromFloat64(13.88)))
	assert.True(t, decimal.NewFromFloat64(13.88).GreaterEqual(decimal.NewFromFloat64(13.88)))
	assert.True(t, decimal.NewFromFloat64(13.88).Greater(decimal.NewFromFloat64(13.11)))
	assert.False(t, decimal.NewFromFloat64(13.88).Less(decimal.NewFromFloat64(13.88)))
	assert.True(t, decimal.NewFromFloat64(13.88).LessEqual(decimal.NewFromFloat64(13.88)))
	assert.True(t, decimal.NewFromFloat64(13.02).Less(decimal.NewFromFloat64(13.11)))
	assert.True(t, decimal.NewFromFloat64(13.02).IsPositive())
	assert.False(t, decimal.NewFromFloat64(-13.02).IsPositive())
	assert.False(t, decimal.NewFromFloat64(-13.02).IsZero())
	assert.False(t, decimal.NewFromFloat64(13.02).IsZero())
	assert.True(t, decimal.NewFromFloat64(13.02).Sub(decimal.NewFromFloat64(13.02)).IsZero())
	assert.True(t, decimal.NewFromFloat64(13.02).Sub(decimal.NewFromFloat64(13.12)).IsNeg())
}

// BenchmarkAddSub-10           53663954                22.47 ns/op            0 B/op          0 allocs/op
func BenchmarkAddSub(b *testing.B) {
	float3 := decimal.NewFromFloat64(3)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		float3.Add(float3)
		float3.Sub(float3)
		decimal.NewFromString("100.000").String()
	}
}

func TestRound(t *testing.T) {
	assert.Equal(t, decimal.NewFromFloat64(13.88).Round(2).String(), "13.88")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Round(1).String(), "13.9")
	assert.Equal(t, decimal.NewFromFloat64(13.44).Round(1).String(), "13.4")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Round(0).Neg().String(), "-14")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Round(0).Neg().Abs().String(), "14")
	assert.Equal(t, decimal.NewFromFloat64(13.88).Round(0).Abs().String(), "14")
	assert.Equal(t, decimal.NewFromFloat64(13.14).Round(1).String(), "13.1")
	assert.Equal(t, decimal.NewFromFloat64(13.15).Round(1).String(), "13.2")
	assert.Equal(t, decimal.NewFromFloat64(13.5).Round(0).String(), "14")
}

func TestCeil(t *testing.T) {
	assert.Equal(t, decimal.NewFromFloat64(13.88).Ceil().String(), "14")
	assert.Equal(t, decimal.NewFromFloat64(13).Ceil().String(), "13")
	assert.Equal(t, decimal.NewFromFloat64(13.0001).Ceil().String(), "14")
	assert.Equal(t, decimal.NewFromFloat64(347.15).Ceil().String(), "348")
}

func TestFloor(t *testing.T) {
	assert.Equal(t, decimal.NewFromFloat64(13.88).Floor().String(), "13")
	assert.Equal(t, decimal.NewFromFloat64(13).Floor().String(), "13")
	assert.Equal(t, decimal.NewFromFloat64(13.0001).Floor().String(), "13")
}
