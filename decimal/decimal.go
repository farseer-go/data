package decimal

import (
	"database/sql/driver"
	"fmt"
	"math"

	govalues "github.com/govalues/decimal"
)

var Zero = Decimal{source: govalues.Zero}
var One = Decimal{source: govalues.One}
var Two = Decimal{source: govalues.Two}
var Float001 = Decimal{source: govalues.MustParse("0.01")}

func NewFromString(s string) Decimal {
	d, _ := govalues.Parse(s)
	return Decimal{source: d}
}
func NewFromFloat64(e float64) Decimal {
	d, _ := govalues.NewFromFloat64(e)
	return Decimal{source: d}
}
func NewFromInt(e int) Decimal {
	d, _ := govalues.New(int64(e), 0)
	return Decimal{source: d}
}
func NewFromInt64(e int64) Decimal {
	d, _ := govalues.New(e, 0)
	return Decimal{source: d}
}
func NewFromInt32(e int32) Decimal {
	d, _ := govalues.New(int64(e), 0)
	return Decimal{source: d}
}

// 在govalues.Decimal的基础上包装一层，用于减少error类型的返回
type Decimal struct {
	source govalues.Decimal
}

func (d Decimal) Add(e Decimal) Decimal {
	r, _ := d.source.Add(e.source)
	return Decimal{source: r}
}

func (d Decimal) AddString(s string) Decimal {
	e, _ := govalues.Parse(s)
	r, _ := d.source.Add(e)
	return Decimal{source: r}
}

func (d Decimal) Sub(e Decimal) Decimal {
	r, _ := d.source.Sub(e.source)
	return Decimal{source: r}
}

func (d Decimal) Mul(e Decimal) Decimal {
	r, _ := d.source.Mul(e.source)
	return Decimal{source: r}
}

func (d Decimal) MulFloat001() Decimal {
	r, _ := d.source.Mul(Float001.source)
	return Decimal{source: r}
}

func (d Decimal) Div(e Decimal) Decimal {
	r, _ := d.source.Quo(e.source)
	return Decimal{source: r}
}

// 四舍五入
func (d Decimal) Round(scale int) Decimal {
	return Decimal{source: d.source.Round(scale)}
}

// Ceil 返回大于或等于 d 的最近整数值。
func (d Decimal) Ceil() Decimal {
	return Decimal{source: d.source.Ceil(0)}
}

// Floor 返回小于或等于 d 的最近整数值。
func (d Decimal) Floor() Decimal {
	return Decimal{source: d.source.Floor(0)}
}

func (d Decimal) Float64() float64 {
	f, _ := d.source.Float64()
	return f
}

// 取反数
func (d Decimal) Neg() Decimal {
	return Decimal{source: d.source.Neg()}
}

// 只取整数部份
func (d Decimal) IntPart() int64 {
	f, _ := d.source.Float64()
	f = math.Floor(f)
	return int64(f)
}

// Truncate 舍去小数部份
func (d Decimal) Truncate() Decimal {
	return Decimal{source: d.source.Floor(0)}
}

// Abs 取绝对值
func (d Decimal) Abs() Decimal {
	return Decimal{source: d.source.Abs()}
}

func (d Decimal) String() string {
	//data := d.removeSufferZero()
	//return *(*string)(unsafe.Pointer(&data))
	return string(d.removeSufferZero())
}

// StringFixed 四舍五入
//
// Example:
//
//	NewFromFloat(0).StringFixed(2) // output: "0.00"
//	NewFromFloat(0).StringFixed(0) // output: "0"
//	NewFromFloat(5.45).StringFixed(0) // output: "5"
//	NewFromFloat(5.45).StringFixed(1) // output: "5.5"
//	NewFromFloat(5.45).StringFixed(2) // output: "5.45"
//	NewFromFloat(5.45).StringFixed(3) // output: "5.450"
//	NewFromFloat(545).StringFixed(-1) // output: "550"
func (d Decimal) StringFixed(pointStatCoinsPoint int) string {
	f, _ := d.source.Round(pointStatCoinsPoint).Float64()
	return fmt.Sprintf("%.*f", pointStatCoinsPoint, f)
}

// true  d > e
// false  other
func (d Decimal) Greater(e Decimal) bool {
	return d.source.Cmp(e.source) > 0
}

// true  d >= e
// false  other
func (d Decimal) GreaterEqual(e Decimal) bool {
	return d.source.Cmp(e.source) >= 0
}

// true  d < e
// false  other
func (d Decimal) Less(e Decimal) bool {
	return d.source.Cmp(e.source) < 0
}

// true  d <= e
// false  other
func (d Decimal) LessEqual(e Decimal) bool {
	return d.source.Cmp(e.source) <= 0
}

// true  d = e
// false  other
func (d Decimal) Equal(e Decimal) bool {
	return d.source.Cmp(e.source) == 0
}

// IsPositive 是否正数
// true  if d > 0
// false otherwise
func (d Decimal) IsPositive() bool {
	return d.source.IsPos()
}

func (d Decimal) IsZero() bool {
	return d.source.IsZero()
}

// IsNeg 是否为负数
//
//	true  if d < 0
//	false otherwise
func (d Decimal) IsNeg() bool {
	return d.source.IsNeg()
}

//****************** 以下是数据库、json等接口需要******************

func (d *Decimal) Scan(value any) error            { return d.source.Scan(value) }
func (d Decimal) Value() (driver.Value, error)     { return d.String(), nil }
func (d *Decimal) UnmarshalJSON(data []byte) error { return d.source.UnmarshalJSON(data) }
func (d Decimal) MarshalJSON() ([]byte, error) {
	text := make([]byte, 0, 26)
	text = append(text, '"')
	text = append(text, d.removeSufferZero()...)
	text = append(text, '"')
	return text, nil
}
func (d *Decimal) UnmarshalText(text []byte) error       { return d.source.UnmarshalText(text) }
func (d Decimal) AppendText(text []byte) ([]byte, error) { return d.source.AppendText(text) }
func (d Decimal) MarshalText() ([]byte, error) {
	data := d.removeSufferZero()
	return data, nil
}
func (d *Decimal) UnmarshalBinary(data []byte) error       { return d.source.UnmarshalBinary(data) }
func (d Decimal) AppendBinary(data []byte) ([]byte, error) { return d.source.AppendBinary(data) }
func (d Decimal) MarshalBinary() ([]byte, error) {
	data := d.removeSufferZero()
	return data, nil
}
func (d *Decimal) UnmarshalBSONValue(typ byte, data []byte) error {
	return d.source.UnmarshalBSONValue(typ, data)
}
func (d Decimal) MarshalBSONValue() (typ byte, data []byte, err error) {
	return d.source.MarshalBSONValue()
}

// 移除小数点末尾的0
func (d Decimal) removeSufferZero() []byte {
	data, _ := d.source.MarshalText()
	// 先确定是否包含小数点
	isHavePoint := false
	for _, d := range data {
		if d == 46 { // 小数点=46
			isHavePoint = true
			break
		}
	}
	// 不包含小数点时，不需要处理，如果末尾是0和小数点，则移除
	for isHavePoint {
		switch data[len(data)-1] {
		case 46: // 小数点=46
			return data[0 : len(data)-1]
		case 48: // 0=46
			data = data[0 : len(data)-1]
		default:
			return data
		}
	}
	return data
}
