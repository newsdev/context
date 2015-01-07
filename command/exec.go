package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/buth/context/backend"
	"github.com/buth/context/crypter"
)

type ExecCommand struct {
	Group, Addr, PrivateKeyFilepath string
	UseEnvironment                  bool
}

func (s *ExecCommand) Run(args []string) int {
	var keyPath, group, template, crypterType, backendType, backendProtocol, backendAddress, backendNamespace string
	flagArgs := flag.NewFlagSet("exec", flag.ContinueOnError)
	flagArgs.StringVar(&backendAddress, "a", ":4001", "backend address")
	flagArgs.StringVar(&backendNamespace, "n", "context", "backend namespace prefix")
	flagArgs.StringVar(&backendProtocol, "protocol", "tcp", "backend protocol")
	flagArgs.StringVar(&backendType, "backend", "etcd", "backend to use")
	flagArgs.StringVar(&crypterType, "crypter", "std", "crypter to use")
	flagArgs.StringVar(&group, "g", "default", "group")
	flagArgs.StringVar(&keyPath, "k", "/etc/context/key", "path to a key file")
	flagArgs.StringVar(&template, "t", "", "cli template")
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

	b, err := backend.NewBackend(backendType, backendNamespace, backendProtocol, backendAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	ecryptedEnv, err := b.GetGroup(group)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	env := make(map[string]string)
	for _, variable := range os.Environ() {
		components := strings.Split(variable, "=")
		env[components[0]] = components[1]
	}

	templateSplit := strings.Split(template, ` `)
	templateArgs := make([]string, 0)
	for variable, encryptedValue := range ecryptedEnv {

		value, err := crypter.ValidateAndDecrypt(encryptedValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		env[variable] = string(value)

		for _, templateComponent := range templateSplit {
			templateArgs = append(templateArgs, strings.Replace(templateComponent, `{}`, variable, -1))
		}
	}

	// Find the expanded path to cmd.
	command, err := exec.LookPath(flagArgs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	commandEnv := make([]string, 0, len(env))
	for key, value := range env {
		commandEnv = append(commandEnv, fmt.Sprintf("%s=%s", key, value))
	}

	commandArgs := make([]string, 0)
	for _, arg := range flagArgs.Args() {
		if arg == `{}` {
			commandArgs = append(commandArgs, templateArgs...)
		} else {

			commandArgs = append(commandArgs, arg)
		}
	}

	if err := syscall.Exec(command, commandArgs, commandEnv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func (s *ExecCommand) Help() string { return "" }

func (s *ExecCommand) Synopsis() string { return "" }
