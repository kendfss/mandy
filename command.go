package mandy

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kendfss/but"
	"github.com/kendfss/iters/slices"
	"github.com/kendfss/oprs"
)

// A Command represents a set of defined flags. The zero value of a Command
// has no name and has ContinueOnError error handling.
//
// Flag names must be unique within a Command. An attempt to define a flag whose
// name is already in use will cause a panic.
type Command struct {
	output      io.Writer
	parent      *Command
	actual      map[string]*Flag
	formal      map[string]*Flag
	Usage       func() string
	Main        func(self *Command) error
	Format      string
	name        string
	URL         string
	children    []*Command
	args        []string
	aliases     []string
	help        helpNode
	parsed      bool
	errorPolicy ErrorPolicy
	lambda      bool // indicates whether the lambda flag was invoked
}

// sortFlags returns the flags as a slice in lexicographical sorted order.
func sortFlags(flags map[string]*Flag) []*Flag {
	result := make([]*Flag, len(flags))
	i := 0
	for _, flag := range flags {
		result[i] = flag
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (c *Command) ChildNames() []string {
	return c.childNames()
}

func (c *Command) Children() []*Command {
	out := make([]*Command, len(c.children))
	ctr := 0
	for _, val := range c.children {
		out[ctr] = val
		ctr++
	}
	return out
}

func (c *Command) childNames() (out []string) {
	for _, child := range c.children {
		out = append(out, child.name)
		out = append(out, child.aliases...)
	}

	return out
}

func (c *Command) AddAlias(args ...string) error {
	blocked := []string{}
	if pcn := c.parent.parent.childNames(); c.parent != nil {
		slices.Sort(pcn)
		pcn = slices.Compact(pcn)
		for _, arg := range args {
			if slices.Contains(pcn, arg) {
				blocked = append(blocked, arg)
			}
		}
	}

	if len(blocked) > 0 {
		return fmt.Errorf("the following args are taken: %v", blocked)
	}

	c.aliases = append(c.aliases, args...)
	return nil
}

// Output returns the destination for usage and error messages. os.Stderr is returned if
// output was not set or was set to nil.
func (c *Command) Output() io.Writer {
	if c.output == nil {
		return os.Stderr
	}
	return c.output
}

// Name returns the name of the flag set.
func (c Command) Name() string {
	return c.name
}

// ErrorPolicy returns the error handling behavior of the flag set.
func (c *Command) ErrorPolicy() ErrorPolicy {
	return c.errorPolicy
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (c *Command) SetOutput(output io.Writer) {
	c.output = output
}

// VisitAll visits the flags in lexicographical order, calling fn for each.
// It visits all flags, even those not set.
func (c *Command) VisitAll(fn func(*Flag)) {
	for _, flag := range sortFlags(c.formal) {
		fn(flag)
	}
}

// Visit visits the flags in lexicographical order, calling fn for each.
// It visits only those flags that have been set.
func (c *Command) VisitSet(fn func(*Flag)) {
	for _, flag := range sortFlags(c.actual) {
		fn(flag)
	}
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func (c *Command) Lookup(name string) *Flag {
	return c.formal[name]
}

// Set sets the value of the named flag.
func (c *Command) Set(name, value string) error {
	flag, ok := c.formal[name]
	if !ok {
		return fmt.Errorf("no such flag -%v", name)
	}
	err := flag.Value.Set(value)
	if err != nil {
		return err
	}
	if c.actual == nil {
		c.actual = make(map[string]*Flag)
	}
	c.actual[name] = flag
	return nil
}

// a string describing the default values of all defined command-line flags in the set.
func (c *Command) Defaults() string {
	return c.usageFlags()
}

// defaultUsage is the default function to print a usage message.
func (c *Command) defaultUsage() string {
	return strings.Join([]string{c.usageHeader(), c.usageFlags(), c.URL}, "\n")
}

func (c Command) usageHeader() string {
	return fmt.Sprintf("usage: %s", c.format())
}

func (c Command) usageFlags() (out string) {
	for _, flag := range c.formal {
		out += "\t" + flag.usage() + "\n"
	}
	return
}

func (c Command) name_() string {
	var names []string
	switch c.parent {
	case nil:
		names = []string{c.name}
	default:
		names = []string{c.parent.name, c.name}
	}
	// return strings.Join(names, " ")
	return strings.Join(names, NameSep)
}

func (c Command) format() (out string) {
	// if isFstr(c.Format) {
	out += "\t" + fmt.Sprintf(c.Format, c.name_())
	for !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	// } else {
	// 	out += c.name_()
	// }
	return out
}

// NFlag returns the number of flags that have been set.
func (c *Command) NFlag() int { return len(c.actual) }

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func (c *Command) Arg(i int) string {
	if i < 0 || i >= len(c.args) {
		return ""
	}
	return c.args[i]
}

func (c Command) Invoked() bool {
	return c.NArg()+c.NFlag() > 0
}

// NArg is the number of arguments remaining after flags have been processed.
func (c *Command) NArg() int { return len(c.args) }

// Args returns the non-flag arguments.
func (c *Command) Args() []string { return c.args }

// Argc returns a channel to the non-flag arguments.
func (c *Command) Argch() chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, arg := range c.args {
			out <- arg
		}
	}()
	return out
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (c *Command) Bool(p *bool, name string, value bool, usage string, short bool) *Flag {
	return c.Var(newBoolValue(value, p), name, usage, short)
}

// Int defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (c *Command) Int(p *int, name string, value int, usage string, short bool) *Flag {
	return c.Var(newIntValue(value, p), name, usage, short)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (c *Command) Int64(p *int64, name string, value int64, usage string, short bool) *Flag {
	return c.Var(newInt64Value(value, p), name, usage, short)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (c *Command) Uint(p *uint, name string, value uint, usage string, short bool) *Flag {
	return c.Var(newUintValue(value, p), name, usage, short)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (c *Command) Uint64(p *uint64, name string, value uint64, usage string, short bool) *Flag {
	return c.Var(newUint64Value(value, p), name, usage, short)
}

// String defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (c *Command) String(p *string, name string, value string, usage string, short bool) *Flag {
	return c.Var(newStringValue(value, p), name, usage, short)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (c *Command) Float64(p *float64, name string, value float64, usage string, short bool) *Flag {
	return c.Var(newFloat64Value(value, p), name, usage, short)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (c *Command) Duration(p *time.Duration, name string, value time.Duration, usage string, short bool) *Flag {
	return c.Var(newDurationValue(value, p), name, usage, short)
}

// Func defines a flag with the specified name and usage string.
// Each time the flag is seen, fn is called with the value of the flag.
// If fn returns a non-nil error, it will be treated as a flag value parsing error.
func (c *Command) Func(fn func(string) error, name, usage string, short bool) *Flag {
	return c.Var(funcValue(fn), name, usage, short)
}

// Check if a command accepts a given flag name
// return the name of the matching flag
// else empty string
func (c *Command) accepts(name string) string {
	for k, v := range c.formal {
		if (v.Short && name == k[:1]) || name == k {
			return k
		}
	}
	return ""
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (c *Command) Var(value Getter, name string, usage string, short bool) *Flag {
	// Flag must not begin "-" or contain "=".
	if strings.HasPrefix(name, "-") {
		panic(c.sprintf("flag %q begins with -", name))
	} else if strings.Contains(name, "=") {
		panic(c.sprintf("flag %q contains =", name))
	}

	// Remember the default value as a string; it won't change.
	flag := &Flag{
		Name:        name,
		Description: usage,
		Value:       value,
		DefValue:    value.String(),
		Short:       short,
	}
	_, alreadythere := c.formal[name]
	if alreadythere {
		var msg string
		if c.name == "" {
			msg = c.sprintf("flag redefined: %s", name)
		} else {
			msg = c.sprintf("%s flag redefined: %s", c.name, name)
		}
		panic(msg) // Happens only if flags are declared with identical names
	}
	if flag.Short {
		for _, other := range c.formal {
			if other.Name != flag.Name && other.Name[0] == flag.Name[0] && other.Short {
				if HelpName == other.Name {
					other.Short = false
					continue
				}
				panic(c.sprintf("Short name collision between %q and %q flags", flag.Name, other.Name))
			}
		}
	}

	if c.formal == nil {
		c.formal = make(map[string]*Flag)
	}
	c.formal[name] = flag

	return flag
}

// sprintf formats the message, prints it to output, and returns it.
func (c *Command) sprintf(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(c.Output(), msg)
	return msg
}

// failf prints to standard error a formatted error and usage message and
// returns the error.
func (c *Command) failf(format string, a ...any) error {
	msg := c.sprintf(format, a...)
	c.usage()
	return errors.New(msg)
}

// usage calls the Usage method for the flag set if one is specified,
// or the appropriate default usage function otherwise.
func (c *Command) usage() {
	if c.Usage == nil {
		c.defaultUsage()
	} else {
		c.Usage()
	}
}

func (c *Command) Visited(f *Flag) bool {
	_, ok := c.actual[f.Name]
	return ok
}

func (c *Command) shortables() (out []*Flag) {
	for _, flag := range c.formal {
		if flag.Short {
			out = append(out, flag)
		}
	}
	return out
}

// Lambdad checks if the command's lambda flag was invoked
func (c *Command) Lambdad() bool {
	return c.lambda
}

// parseTrailer finds param-terminated-bool-sequences like "-abcd e"
func (c *Command) parseTrailer(f *Flag) (*Command, bool, error) {
	// dashes := oprs.Ternary(f.ShortingPolicy == )
	c.args = append([]string{"--" + f.Name}, c.args...)
	return c.parseOne()
}

func abbrev(s string) string {
	if len(s) > 10 {
		return string([]rune(s)[:7]) + "..."
	}
	return s
}

type argTk uint64

const (
	aknull argTk = 1 << iota
	akchild
	akbool
	aknon
	akpred
	aksucc
	akclassic
	akshort   // - prefixed flag
	aklong    // -- prefixed flag
	akfree    // unflagged argument
	akeoflags // "-"
	akeoargs  // "--"
)

func (cmd *Command) expandArgs(shorts map[string]string, args ...string) []string {
	var expandedArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
			if strings.Contains(arg, "=") {
				// Short option with value
				optionWithValue := strings.SplitN(arg[1:], "=", 2)
				shortOpt := optionWithValue[0]
				if fullOpt, ok := shorts[shortOpt]; ok {
					expandedArgs = append(expandedArgs, "--"+fullOpt+"="+optionWithValue[1])
				} else {
					// Unknown short option, keep it as is
					expandedArgs = append(expandedArgs, arg)
				}
			} else {
				// Expand short option
				shortOpts := arg[1:]
				for j := 0; j < len(shortOpts); j++ {
					shortOpt := string(shortOpts[j])
					if fullOpt, ok := shorts[shortOpt]; ok {
						expandedArgs = append(expandedArgs, "--"+fullOpt)
					} else {
						// Unknown short option, keep it as is
						expandedArgs = append(expandedArgs, "-"+string(shortOpt))
					}
				}

				// Check if the last short option has more terms
				if i+1 < len(args) && len(shortOpts) > 1 {
					expandedArgs[len(expandedArgs)-1] += "=" + args[i+1]
					i++
				}
			}
		} else {
			expandedArgs = append(expandedArgs, arg)
		}
	}

	return expandedArgs
}

func (c *Command) parseOne() (*Command, bool, error) {
	if len(c.args) == 0 {
		return nil, false, nil
	}
	arg := c.args[0]
	c.args = c.args[1:]
	// Check if it's a flag-value pair
	if strings.Contains(arg, "=") {
		parts := strings.SplitN(arg, "=", 2)
		flagName := parts[0]
		flagValue := parts[1]
		// Find the flag in the command's flag set
		flag := c.formal[c.accepts(flagName)]
		if flag == nil {
			return nil, false, fmt.Errorf("unknown flag: %s", flagName)
		}
		// Check if the flag has a value type other than bool
		if !flag.Value.IsBool() {
			if err := flag.Value.Set(flagValue); err != nil {
				return nil, false, fmt.Errorf("invalid value for flag %s: %s", flagName, flagValue)
			}
		} else {
			return nil, false, fmt.Errorf("unexpected value for boolean flag: %s", flagName)
		}
		return nil, true, nil
	}
	// Check if it's a long flag
	if strings.HasPrefix(arg, "--") {
		flagName := strings.TrimPrefix(arg, "--")
		flag := c.formal[c.accepts(flagName)]
		if flag == nil {
			return nil, false, fmt.Errorf("unknown flag: %s", flagName)
		}
		// Check if the flag is a bool flag

		if f, ok := flag.Value.Get().(boolFlag); f != nil && ok {
			flag.Value.Set("true")
		} else {
			return nil, false, fmt.Errorf("missing value for non-boolean flag: %s", flagName)
		}
		return nil, true, nil
	}
	// Check if it's a short flag or a shorthand for a long flag
	if strings.HasPrefix(arg, "-") {
		flagNames := strings.TrimPrefix(arg, "-")
		for i, flagName := range flagNames {
			flag := c.formal[c.accepts(string(flagName))]
			if flag == nil {
				return nil, false, fmt.Errorf("unknown flag: %s", string(flagName))
			}
			// Check if the flag is a bool flag
			if flag.Value.IsBool() {
				flag.Value.Set("true")
			} else if i == len(flagNames)-1 {
				// Last term is assumed to be the value for non-boolean flag
				if len(c.args) == 0 {
					return nil, false, fmt.Errorf("missing value for non-boolean flag: %s", string(flagName))
				}
				flag.Value.Set(c.args[0])
				c.args = c.args[1:]
			} else {
				return nil, false, fmt.Errorf("unexpected value for boolean flag: %s", string(flagName))
			}
		}
		return nil, true, nil
	}
	return nil, false, nil
}

// MustParse parses flag definitions from the argument list
func (c *Command) MustParse() {
	c.Handle(c.Parse())
}

// WarnIf prints help and exits if help is needed
func (c *Command) WarnIf(b bool, fmtArgs ...any) {
	if !b {
		if len(fmtArgs) > 0 {
			msg := fmtArgs[0].(string)
			if []rune(msg)[len([]rune(msg))-1] != '\n' {
				msg += "\n"
			}
			c.Warnf(msg, fmtArgs[1:]...)
		}
	}
}

// HelpIf prints help and exits if help is needed
func (c *Command) HelpIf(b bool, fmtArgs ...any) {
	if b {
		if len(fmtArgs) > 0 {
			msg := fmtArgs[0].(string)
			if []rune(msg)[len([]rune(msg))-1] != '\n' {
				msg += "\n"
			}
			fmt.Printf(msg, fmtArgs[1:]...)
		}
		c.PrintHelp()
	}
}

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the Command
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
// func (c *Command) Parse(arguments []string) error {
func (c *Command) Parse(args ...string) error {
	defer c.setparsed()
	switch {
	case c.parent != nil:
		c.args = c.parent.args[1:]
	case len(args) != 0:
		c.args = args
	default:
		c.args = os.Args[1:]
	}
	for {
		child, seen, err := c.parseOne()
		if seen {
			continue
		}
		if child != nil {
			return child.Parse()
		}
		if err == nil {
			break
		}
		c.Handle(err)
	}
	return nil
}

func (c *Command) setparsed() {
	c.parsed = true
}

// Parsed reports whether c.Parse has been called.
func (c Command) Parsed() bool {
	return c.parsed
}

func (c *Command) SetHelpFlag(name string, short bool) (out *Flag) {
	delete(c.formal, HelpName)
	p := new(bool)
	out = c.Var(newBoolValue(false, p), name, "print this message", short)
	HelpName = name
	return
}

// func (c *Command) HelpFlag() *Flag {}

// NewCommand returns a new, empty flag set with the specified name and
// error handling property. If the name is not empty, it will be printed
// in the default usage message and in error messages.
func NewCommand(name string, errorPolicy ErrorPolicy) *Command {
	c := &Command{
		name:        name,
		errorPolicy: errorPolicy,
		Format:      "%s [options] [args...]",
		URL:         EnvUrl(name),
	}
	if name != HelpName {
		p := new(bool)
		c.Var(newBoolValue(false, p), HelpName, "print this message", true)
		c.Usage = c.defaultUsage
	}
	return c
}

// NewSubCommand returns a new, empty flag set with the specified name.
// If the name is not empty, it will be printed
// in the default usage message and in error messages.
// The ErrorPolicy will be inherited from the command.
// If the name is set to "help" it will not have a help flag
func (c *Command) NewChild(name string) *Command {
	s := NewCommand(name, c.errorPolicy)
	s.parent = c
	s.URL = c.URL
	c.children = append(c.children, s)
	return s
}

func (c *Command) first() *Command {
	// current := c
	for current, parent := c, c.parent; current.parent != nil; current, parent = parent, parent.parent {
		if parent.parent == nil {
			return parent
		}
	}
	return c
}

// Run a command's "Main" attribute on a specific set of arguments
// Overrides os.Args usage
// Returns ErrNilMain if command.Main is nil.
func (c *Command) Execute(args ...string) error {
	c.args = args

	err := c.Parse()
	if err == nil {
		if c.Main != nil {
			return c.Main(c)
		}
		return ErrNilMain
	}

	return err
}

// Init sets the name and error handling property for a flag set.
// By default, the zero Command uses an empty name and the
// ContinueOnError error handling policy.
func (c *Command) Init(name string, errorPolicy ErrorPolicy) {
	c.name = name
	c.errorPolicy = errorPolicy
}

// a bash-value-safe wrapper on os.Exit
// appends a \n to msg if msg's non-empty and not \n terminated
// always writes to stderr
func (c Command) Exit(msg string, code uint8) {
	c.Warn(but.New(msg))
	os.Exit(int(code))
}

// Print the Usage() text and exit with error code #1
func (c Command) PrintHelp() {
	c.Exit(c.Usage(), 1)
}

// Behave as consistent with the chosen error handling method
// this does nothing if the error is nil (
//
//	so you can save yourself some of the hassle of handling
//	errors manually unless you're handling a special case
//
// ).
func (c Command) Handle(err error) {
	if err != nil {
		switch c.errorPolicy {
		case ContinueOnError:
			fmt.Fprintln(os.Stderr, err)
		case ExitOnError:
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		case PanicOnError:
			panic(err)
		default:
			panic("unrecognized error policy")
		}
	}
}

// print an error to stderr if, and only if, it is not nil
func (c Command) Warn(err error) {
	if err == nil {
		return
	}
	msg := err.Error()
	if len(msg) > 0 {
		if msg[len(msg)-1] != '\n' {
			msg += "\n"
		}
		os.Stderr.WriteString(msg)
	}
	// if err != nil {
	// 	os.Stderr.WriteString(err.Error() + "\n")
	// }
}

func (c Command) Warnf(msg string, args ...any) {
	if len(msg) == 0 {
		c.Warn(nil)
	} else if len(args) == 0 {
		c.Warn(but.New(msg))
	} else {
		c.Warn(but.New(msg, args...))
	}
}

// Check if command is receiving input via stdin
func (c *Command) Receiving() bool {
	stat, err := os.Stdin.Stat()
	c.Handle(err)
	return stat.Size() > 0
}

// Infer whether or not the user needs help
// Checks if
//
//	 parsed
//		the help flag was invoked
//		or no args/flags were invoked
//
// hacker note:
//
//	herein lies a panic that will trigger if you unset the default help flag
func (c Command) HelpWorthy() bool {
	_, defined := c.formal[HelpName]
	but.MustBool(defined, "help flag %q is undefined for this command", HelpName)

	_, used := c.actual[HelpName]

	noFlags := c.NFlag() == 0
	noArgs := c.NArg() == 0

	return c.Parsed() && (used || (noFlags && noArgs && !c.Receiving()))
}

// Similar to HelpNeeded but does not check if flags or args have been set
// Checks if
//
//	 parsed
//		the help flag was invoked
//
// hacker note:
//
//	herein lies a panic that will trigger if you unset the default help flag
func (c Command) HelpNeeded() bool {
	_, defined := c.formal[HelpName]
	// but.Must(defined, "help flag %q is undefined for this command", HelpName)
	// println("defined", defined)
	but.MustBool(defined, errUndefinedHelp.Fmt(HelpName))

	_, used := c.actual[HelpName]

	return c.Parsed() && used
}

const errUndefinedHelp but.Note = "help flag %q is undefined for this command"

// Deprecated
// Checks if
//
//	 HelpNeeded
//		or HelpWorthy
//
// hacker note:
//
//	hereinlies a panic that will trigger if you unset the default help flag
func (c *Command) HelpWanted() bool {
	return c.HelpNeeded() || c.HelpWorthy()
}

// derive the keys of a map
func keys[K comparable, V any](m map[K]V) (out []K) {
	for k := range m {
		out = append(out, k)
	}
	return
}

// derive the values of a map
func values[K comparable, V any](m map[K]V) (out []V) {
	for k := range m {
		out = append(out, m[k])
	}
	return
}

func filter[T any](f func(T) bool, args ...T) (out []T) {
	for _, x := range args {
		if f(x) {
			out = append(out, x)
		}
	}
	return
}

func isFstr(s string) bool {
	return filter(oprs.Method(s, strings.Contains), "%s", "%v", "%#v") != nil
}

// derive a url from the $REPO_HOST and $DEVELOPER environment variables
// name refers to the name of the cli/command
func EnvUrl(name string) string {
	var (
		repoHost = os.Getenv("REPO_HOST")
		devName  = os.Getenv("DEVELOPER")
	)
	out, err := url.JoinPath(repoHost, devName, name)
	but.Must(err)
	return out
}
