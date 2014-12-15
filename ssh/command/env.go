package command

import (
	"fmt"
	"io"

	"github.com/nytinteractive/context/backend"
	"github.com/nytinteractive/context/crypter"
)

type EnvCommand struct {
	Backend backend.Backend
	Crypter crypter.Crypter
}

func (e *EnvCommand) Run(args []string, env map[string]string, stdout io.Writer) int {

	if len(args) != 1 {
		return 1
	}

	groupEnv, err := e.Backend.GetGroup(env[`GROUP`])
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	for variable, valueEncryptedBytes := range groupEnv {

		valueBytes, err := e.Crypter.ValidateAndDecrypt(valueEncryptedBytes)
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 1
		}

		fmt.Fprintf(stdout, "%s=%s\n", variable, string(valueBytes))
	}

	return 0
}
