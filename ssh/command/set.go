package command

import (
	"fmt"
	"io"

	"github.com/nytinteractive/context/backend"
	"github.com/nytinteractive/context/crypter"
)

type SetCommand struct {
	Backend backend.Backend
	Crypter crypter.Crypter
}

func (s *SetCommand) Run(args []string, env map[string]string, stdout io.Writer) int {

	if len(args) != 2 {
		return 1
	}

	variable := args[1]
	value, ok := env[variable]
	if !ok {
		return 1
	}

	valueCryptedBytes, err := s.Crypter.EncryptAndSign([]byte(value))
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	if err := s.Backend.SetVariable(env[`GROUP`], variable, valueCryptedBytes); err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	return 0
}
