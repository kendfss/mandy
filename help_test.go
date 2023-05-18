package mandy

import (
	"fmt"
	"os"
	"testing"
)

func TestHelpFile(t *testing.T) {
	name, Repo := "bckp", "https://github.com/kendfss/bckp"
	help := helpMessage{
		{depth: 0, text: fmt.Sprintf("%s: bundle, separate, and organize with archives or directories", name)},
		{
			depth: 1,
			text:  fmt.Sprintf("Usage:"),
			children: helpMessage{
				{
					depth: 1,
					text:  txt(false, "pop"),
					children: helpMessage{
						{
							depth: 1,
							text:  txt(false, "nest"),
						},
						{
							depth: 1,
							text:  txt(false, "verbose"),
						},
						{
							depth: 1,
							text:  txt(false, "discard"),
						},
						{
							depth: 1,
							text:  txt(false, "zip"),
						},
					},
				},
				{
					depth: 1,
					text:  txt(false, "put"),
					children: helpMessage{
						{
							depth: 1,
							text:  txt(false, "nest"),
						},
						{
							depth: 1,
							text:  txt(false, "verbose"),
						},
						{
							depth: 1,
							text:  txt(false, "discard"),
						},
						{
							depth: 1,
							text:  txt(false, "zip"),
						},
					},
				},
			},
		},
		// {depth: 2, text: Repo},
		{depth: 0, text: fmt.Sprintf("more info @:\n\t\"man %s\"\n\t%s", os.Args[0], Repo)},
	}

	println(help)

	println("....")

	for _, node := range help {
		println(node.String())
	}
}
