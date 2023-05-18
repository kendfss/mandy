package mandy

// package main

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	defaultIndent = "\t"
)

var (
	HelpName = "help"
	NameSep  = " "
)

type (
	Item[T fmt.Stringer] struct {
		Value    T         `json:"value"`
		Children []Item[T] `json:"children"`
	}

	helpNode struct {
		text     string
		children helpMessage
		depth    int
	}

	flagHelp struct {
		desc  string
		def   string
		name  string
		short bool
	}

	helpMessage []helpNode
)

func (i Item[T]) String() string {
	out := i.Value.String()
	for _, child := range i.Children {
		out += child.String()
	}
	return out
}

func (m helpMessage) String() string {
	out := ""
	for _, node := range m {
		out += node.String()
	}
	return out
}

func (hn helpNode) Resolved() helpNode {
	n := helpNode{
		depth:    hn.depth,
		text:     hn.text,
		children: hn.children,
	}

	for i, child := range n.children {
		n.children[i].depth += n.depth
		n.children[i] = child.Resolved()
	}
	return n
}

func (hn helpNode) String() string {
	n := hn.Resolved()
	msg := strings.Repeat(defaultIndent, n.depth) + n.text
	for _, child := range n.children {
		msg += fmt.Sprintf("\n%s", reindent(strings.Repeat(defaultIndent, n.depth+child.depth)+child.String(), n.depth))
	}

	return msg
}

func (n helpNode) repr(boost int) string {
	msg := strings.Repeat(defaultIndent, boost+n.depth) + n.text
	for _, child := range n.children {
		// boost :=
		msg += fmt.Sprintf("\n%s", child.repr(boost+n.depth))
	}
	return msg
}

// func newHelpNode(indent, text, children) *helpNode {
// 	return
// }

func txt[T any](val T, msg string) string {
	out := fmt.Sprintf("%s [default: %v]", msg, val)
	return fmt.Sprintf("-%s\n\t%s\t%s\n", msg, reflect.TypeOf(val).Name(), out)
}

func NewItem[T fmt.Stringer](val T, children ...T) Item[T] {
	kids := []Item[T]{}
	for _, child := range children {
		kids = append(kids, NewItem(child))
	}
	if len(kids) == 0 {
		kids = *new([]Item[T])
	}
	return Item[T]{
		Value:    val,
		Children: kids,
	}
}

func reindent(orig string, depth int) string {
	lines := strings.Split(orig, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(defaultIndent, depth) + line
	}
	return strings.Join(lines, "\n")
}
