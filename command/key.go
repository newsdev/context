package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/buth/context/crypter"
)

type KeyCommand struct {
	Group, Addr, PrivateKeyFilepath string
	UseEnvironment                  bool
}

func (s *KeyCommand) Run(args []string) int {
	var keyPath, crypterType string
	flagArgs := flag.NewFlagSet("key", flag.ContinueOnError)
	flagArgs.StringVar(&crypterType, "crypter", "std", "crypter to use")
	flagArgs.StringVar(&keyPath, "k", "/etc/context/key", "path to save the key file to")
	if err := flagArgs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Generate a new key of the given type.
	key, err := crypter.NewKey(crypterType)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Create a new file. This will wipe out any existing file (if we can
	// write to it) and set permissions to 666.
	out, err := os.Create(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Set more restrictive permissions on the file *before* we write to it.
	if err := out.Chmod(0600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if _, err := out.Write(key); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := out.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func (s *KeyCommand) Help() string { return "" }

func (s *KeyCommand) Synopsis() string { return "" }
