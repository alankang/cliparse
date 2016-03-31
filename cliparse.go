package cliparse

import (
	"flag"
	"fmt"
	"os"
)

type Cmd struct {
	name, desc string
	helpSum    string
	flag.FlagSet
	subCmds map[string]*Cmd
	run     func(cmd *Cmd) error
}

var rootCmd = New(os.Args[0], "", "", nil)

func New(name, desc, helpSum string, run func(cmd *Cmd) error) *Cmd {
	c := &Cmd{
		name:    name,
		desc:    desc,
		helpSum: helpSum,
		run:     run,
	}
	c.FlagSet.Init(name, flag.ExitOnError)
	c.FlagSet.Usage = c.usage
	return c
}

func RootCmd() *Cmd {
	return rootCmd
}

func (c *Cmd) RegisterSubCmds(subCmd *Cmd) {
	if len(subCmd.name) == 0 {
		panic("sub command has no name")
	}
	if _, ok := c.subCmds[subCmd.name]; ok {
		panic(fmt.Sprintf("sub command '%s' redefined", subCmd.name))
	}
	if c.subCmds == nil {
		c.subCmds = make(map[string]*Cmd)
	}
	c.subCmds[subCmd.name] = subCmd
}

func (c *Cmd) Parse(args []string) *Cmd {
	if err := c.FlagSet.Parse(args); err != nil {
		return nil
	}
	if c.subCmds == nil {
		return c
	}
	if c.FlagSet.NArg() > 0 {
		sc := c.FlagSet.Arg(0)
		if subc, ok := c.subCmds[sc]; !ok {
			fmt.Printf("'%s' not recognized.\n", sc)
			return nil
		} else {
			return subc.Parse(c.FlagSet.Args()[1:])
		}
	}
	return c
}

func Parse() *Cmd {
	return rootCmd.Parse(os.Args[1:])
}

func (c *Cmd) usage() {
	fmt.Printf("Usage for '%s':\n", c.name)
	if len(c.helpSum) > 0 {
		fmt.Println(c.helpSum)
	}
	if len(c.subCmds) > 0 {
		fmt.Printf("Sub commands:\n")
		for name, sc := range c.subCmds {
			fmt.Printf("  %s\n\t%s\n", name, sc.desc)
		}
		fmt.Println()
	}

	var nflags int
	c.FlagSet.VisitAll(func(f *flag.Flag) {
		nflags++
	})

	if nflags > 0 {
		fmt.Printf("Options:\n")
		c.FlagSet.PrintDefaults()
	}
}

func (c *Cmd) Run() error {
	if c.run == nil {
		c.usage()
		return nil
	}
	return c.run(c)
}

func (c *Cmd) Name() string {
	return c.name
}
