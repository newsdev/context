package command

import (
	"fmt"
	"io"

	"github.com/nytinteractive/context/backend"
	"github.com/nytinteractive/context/crypter"
)

type UnsetCommand struct {
	Backend backend.Backend
	Crypter crypter.Crypter
}

func (u *UnsetCommand) Run(args []string, env map[string]string, stdout io.Writer) int {

	if len(args) != 2 {
		return 1
	}

	if err := u.Backend.RemoveVariable(env[`GROUP`], args[1]); err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	return 0
}
