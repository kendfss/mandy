package mandy

import (
	"fmt"
	"reflect"
)

// type FlagSet map[string]*Flag

// A Flag represents the state of a flag.
type Flag struct {
	Name        string // name as it appears on command line
	Description string // outline of the flag's behaviour
	DefValue    string // default value (as text); for usage message
	Short       bool   // whether or not the flag can be referenced by abbreviation
	Value       Getter // value as set
	// Value       Value  // value as set
	// visited bool
}

// Eq checks if a flag has a given value
func (f *Flag) Eq(arg any) bool {
	return reflect.ValueOf(f.Value.Get()).Equal(reflect.ValueOf(arg))
}

// func (f *Flag) Visited() bool {
// 	return f.visited
// }

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(flag *Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(flag.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	return value == z.Interface().(Value).String()
}

// UnquoteDescription extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func UnquoteDescription(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Description
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = "value"
	switch flag.Value.(type) {
	case boolFlag:
		name = ""
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}

func (f Flag) help() flagHelp {
	return flagHelp{
		desc:  f.Description,
		def:   f.DefValue,
		short: f.Short,
		name:  f.Name,
	}

	// return fmt.Sprintf(
	// 	"%s\n"
	// )
}

func (f Flag) usage() (out string) {
	if f.Short {
		out += fmt.Sprintf("-%c, --%s", f.Name[0], f.Name)
	} else {
		out += "--" + f.Name
	}
	out += fmt.Sprintf("\t%s [default: %s]", f.Description, f.DefValue)
	return
}

// UnquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func UnquoteUsage(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Description
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = "value"
	switch flag.Value.(type) {
	case boolFlag:
		name = ""
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}
