package cliparse

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type Cmd struct {
	name, desc string
	// help summary for the cmd, displayed before sub cmds and options list
	helpSum string
	flag.FlagSet
	subCmds   map[string]*Cmd
	dftSubCmd *Cmd
	run       func(cmd *Cmd) error
	output    io.Writer
}

var rootCmd = New(os.Args[0], "", "", nil)

func New(name, desc, helpSum string, run func(cmd *Cmd) error) *Cmd {
	c := &Cmd{
		name:    name,
		desc:    desc,
		helpSum: helpSum,
		run:     run,
		output:  os.Stdout,
	}
	c.FlagSet.Init(name, flag.ContinueOnError)
	c.FlagSet.Usage = c.usage
	return c
}

func RootCmd(run func(cmd *Cmd) error) *Cmd {
	rootCmd.run = run
	return rootCmd
}

func (c *Cmd) SetOutput(out io.Writer) {
	c.output = out
}

func (c *Cmd) registerSubCmd(subCmd *Cmd, dft bool) {
	if len(subCmd.name) == 0 {
		panic("Sub command has no name")
	}
	if c.dftSubCmd != nil && dft {
		panic(fmt.Sprintf("More than one default sub command defined for %s", c.name))
	}
	if _, ok := c.subCmds[subCmd.name]; ok {
		panic(fmt.Sprintf("Sub command '%s' redefined", subCmd.name))
	}
	if c.subCmds == nil {
		c.subCmds = make(map[string]*Cmd)
	}
	c.subCmds[subCmd.name] = subCmd
	if dft {
		c.dftSubCmd = subCmd
	}
}

func (c *Cmd) RegisterDftSubCmd(subCmd *Cmd) {
	c.registerSubCmd(subCmd, true)
}

func (c *Cmd) RegisterSubCmds(subCmds ...*Cmd) {
	for _, sc := range subCmds {
		c.registerSubCmd(sc, false)
	}
}

func (c *Cmd) Parse(args []string) (*Cmd, error) {
	if err := c.FlagSet.Parse(args); err != nil {
		return nil, err
	}
	if c.FlagSet.NArg() > 0 {
		if c.subCmds == nil {
			return c, nil
		}
		sc := c.FlagSet.Arg(0)
		if subc, ok := c.subCmds[sc]; !ok {
			err := fmt.Errorf("Sub command '%s' not recognized.\n", sc)
			fmt.Fprintf(c.output, "%v\n", err)
			return nil, err
		} else {
			return subc.Parse(c.FlagSet.Args()[1:])
		}
	} else {
		if c.dftSubCmd != nil {
			return c.dftSubCmd, nil
		}
	}
	return c, nil
}

func Parse() (*Cmd, error) {
	return rootCmd.Parse(os.Args[1:])
}

func (c *Cmd) usage() {
	fmt.Fprintf(c.output, "Usage for '%s':\n", c.name)
	if len(c.helpSum) > 0 {
		fmt.Fprintf(c.output, "%s\n", c.helpSum)
	}
	if len(c.subCmds) > 0 {
		fmt.Fprintf(c.output, "Sub commands:\n")
		for name, sc := range c.subCmds {
			dft := ' '
			if name == c.dftSubCmd.name {
				dft = '*'
			}
			fmt.Fprintf(c.output, "%c %s\n\t%s\n", dft, name, sc.desc)
		}
		fmt.Fprintf(c.output, "\n")
	}

	var nflags int
	c.FlagSet.VisitAll(func(f *flag.Flag) {
		nflags++
	})

	if nflags > 0 {
		fmt.Fprintf(c.output, "Options:\n")
		c.FlagSet.PrintDefaults()
	}
}

func (c *Cmd) Run() error {
	if c.run == nil {
		err := fmt.Errorf("Command '%s' not implemented.", c.name)
		fmt.Fprintf(c.output, "%v\n", err)
		return err
	}
	return c.run(c)
}

func (c *Cmd) Name() string {
	return c.name
}
