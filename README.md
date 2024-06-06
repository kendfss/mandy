mandy
---

A mildly opinionated command line parser for Go. Generally it's intended for personal projects, but it comes with a `Production` boolean value so that some of the things that would only make sense on your system do not fall apart on other people's any more than they would on yours.

It's influenced by python's `argparse` package, so it allows
- concatenating abbreviations
- `"--" != "-"`

It also supports
- Configuration
- subcommands

To make the most of it out of the box you should set the following environment variables 




todo
- implement generic `newValue[T](ptr *T, val T, desc string, short bool)`
- help should be a sub-command
- `Command.SubCommand` method
- Settings interface: marshall, unmarshall
- `Format` should be a private method
