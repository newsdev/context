package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"code.google.com/p/gopass"
	"github.com/buth/context/backend"
	"github.com/buth/context/crypter"
)

type SetCommand struct {
	Group, Addr, PrivateKeyFilepath string
	UseEnvironment                  bool
}

func (s *SetCommand) Run(args []string) int {
	var keyPath, group, crypterType, backendType, backendProtocol, backendAddress, backendNamespace string
	flagArgs := flag.NewFlagSet("set", flag.ContinueOnError)
	flagArgs.StringVar(&backendAddress, "a", "http://127.0.0.1:4001", "backend address")
	flagArgs.StringVar(&backendNamespace, "n", "context", "backend namespace prefix")
	flagArgs.StringVar(&backendProtocol, "protocol", "tcp", "backend protocol")
	flagArgs.StringVar(&backendType, "backend", "etcd", "backend to use")
	flagArgs.StringVar(&crypterType, "crypter", "std", "crypter to use")
	flagArgs.StringVar(&group, "g", "default", "group")
	flagArgs.StringVar(&keyPath, "k", "/etc/context/key", "path to a key file")
	if err := flagArgs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Check the status of the secret file.
	stat, err := os.Stat(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Only proceed if the running user is the only user that can read the
	// secret.
	if mode := stat.Mode(); mode != 0600 && mode != 0400 {
		fmt.Fprintln(os.Stderr, "incorrect file mode for key")
		return 1
	}

	// Attempt to read the entire content of the secret file.
	file, err := os.Open(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	key, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if err := file.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}

	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Use the key to create a new crypter of the given type.
	crypter, err := crypter.NewCrypter(crypterType, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	b, err := backend.NewBackend(backendType, backendNamespace, backendAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, variable := range flagArgs.Args() {

		var value string
		if envValue := os.Getenv(variable); s.UseEnvironment && envValue != "" {
			value = envValue
		} else {

			// Get the value from user input.
			inputValue, err := gopass.GetPass(fmt.Sprintf("%s=", variable))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}

			value = inputValue
		}

		ecryptedValue, err := crypter.EncryptAndSign([]byte(value))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		if err := b.SetVariable(group, variable, ecryptedValue); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	return 0
}

func (s *SetCommand) Help() string { return "" }

func (s *SetCommand) Synopsis() string { return "" }
