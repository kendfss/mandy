package mandy

import (
	"strconv"
	"time"
)

/*

import "reflect"

typeMap := map[reflect.Type]
type ArgType interface {
	~bool | ~int | ~int64 | ~uint | ~uint64 | ~float64 | ~string | ~func(string)error
}

type ValueType interface {
	~boolValue | ~intValue | ~int64Value | ~uintValue | ~uint64Value | ~stringValue | ~float64Value | ~durationValue | ~funcValue
}

func newValue[T ArgType](val T, p *T) ValueType {
	*p = val
	return (*vt[T])(p) // where vt is a map[type]type
}
*/

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBool() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
//
// Set is called once, in command line order, for each flag present.
// The flag package may call the String method with a zero-valued receiver,
// such as a nil pointer.
type Value interface {
	String() string
	Set(string) error
	IsBool() bool
}

// Getter is an interface that allows the contents of a Value to be retrieved.
// It wraps the Value interface, rather than being part of it, because it
// appeared after Go 1 and its compatibility rules. All Value types provided
// by this package satisfy the Getter interface, except the type used by Func.
type Getter interface {
	Value
	Get() any
}

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() any       { return bool(*b) }
func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }
func (b *boolValue) IsBool() bool   { return true }

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	Value
	IsBool() bool
}

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() any       { return int(*i) }
func (i *intValue) String() string { return strconv.Itoa(int(*i)) }
func (b *intValue) IsBool() bool   { return false }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() any       { return int64(*i) }
func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }
func (b *int64Value) IsBool() bool   { return false }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() any       { return uint(*i) }
func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }
func (b *uintValue) IsBool() bool   { return false }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() any       { return uint64(*i) }
func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }
func (b *uint64Value) IsBool() bool   { return false }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any       { return string(*s) }
func (s *stringValue) String() string { return string(*s) }
func (b *stringValue) IsBool() bool   { return false }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = numError(err)
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() any       { return float64(*f) }
func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }
func (b *float64Value) IsBool() bool   { return false }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		err = errParse
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() any       { return time.Duration(*d) }
func (d *durationValue) String() string { return (*time.Duration)(d).String() }
func (b *durationValue) IsBool() bool   { return false }

// -- function Value
type funcValue func(string) error

func (f funcValue) Set(s string) error { return f(s) }
func (f funcValue) String() string     { return "" }
func (f funcValue) Get() any           { return f }
func (b funcValue) IsBool() bool       { return false }
