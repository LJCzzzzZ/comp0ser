package cmd

import "strings"

type Cmd struct {
	Bin  string
	Args []string

	Inputs  []string
	Outputs []string
}

func (c *Cmd) String() string {
	return c.Bin + " " + strings.Join(c.Args, " ")
}
