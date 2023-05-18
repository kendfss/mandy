package mandy

/*
	Package mandy implements command-line flag parsing.

	Usage

	Define flags using mandy.String(), Bool(), Int(), etc.

	This declares an integer flag, -n, stored in the pointer nFlag, with type *int:
		import "mandy"
		var nFlag = mandy.Int("n", 1234, "help message for flag n")
	If you like, you can bind the flag to a variable using the Var() functions.
		var flag int
		func init() {
			mandy.Int(&flag, "flagname", 1234, "help message for flagname")
		}
	Or you can create custom flags that satisfy the Value interface (with
	pointer receivers) and couple them to flag parsing by
		mandy.Var(&flagVal, "name", "help message for flagname", true)
	For such flags, the default value is just the initial value of the variable.

	After all flags are defined, call
		mandy.Parse()
	to parse the command line into the defined flags.

	Flags may then be used directly. If you're using the flags themselves,
	they are all pointers; if you bind to variables, they're values.
		fmt.Println("ip has value ", *ip)
		fmt.Println("flag has value ", flag)

	After parsing, the arguments following the flags are available as the
	slice flag.Args() or individually as mandy.Arg(i).
	The arguments are indexed from 0 through mandy.NArg()-1.

	Command line flag syntax

	The following forms are permitted:

		-f
		--flag

		-f=x
		--flag=x

		-f x  // non-boolean flags only
		--flag x  // non-boolean flags only

	One or two minus signs may be used; they are equivalent.
	The last form is not permitted for boolean flags because the
	meaning of the command
		cmd -x *
	where * is a Unix shell wildcard, will change if there is a file
	called 0, false, etc. You must use the -flag=false form to turn
	off a boolean flag.

	Flag parsing stops just before the first non-flag argument
	("-" is a non-flag argument) or after the terminator "--".

	Integer flags accept 1234, 0664, 0x1234 and may be negative.
	Boolean flags may be:
		1, 0, t, f, T, F, true, false, TRUE, FALSE, True, False, y, n, Y, N, yes, no, Yes, No, YES, NO
	Duration flags accept any input valid for time.ParseDuration.

	The default set of command-line flags is controlled by
	top-level functions.  The Command type allows one to define
	independent sets of flags, such as to implement subcommands
	in a command-line interface. The methods of Command are
	analogous to the top-level functions for the command-line
	flag set.
*/
