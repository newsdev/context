package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/buth/context/backend"
)

type UnsetCommand struct{}

func (s *UnsetCommand) Run(args []string) int {
	var group, backendType, backendProtocol, backendAddress, backendNamespace string
	flagArgs := flag.NewFlagSet("unset", flag.ContinueOnError)
	flagArgs.StringVar(&backendAddress, "a", ":4001", "backend address")
	flagArgs.StringVar(&backendNamespace, "n", "context", "backend namespace prefix")
	flagArgs.StringVar(&backendProtocol, "protocol", "tcp", "backend protocol")
	flagArgs.StringVar(&backendType, "backend", "etcd", "backend to use")
	flagArgs.StringVar(&group, "g", "default", "group")
	if err := flagArgs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	b, err := backend.NewBackend(backendType, backendNamespace, backendProtocol, backendAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, variable := range flagArgs.Args() {
		if b.RemoveVariable(group, variable); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	return 0
}

func (s *UnsetCommand) Help() string { return "" }

func (s *UnsetCommand) Synopsis() string { return "" }
