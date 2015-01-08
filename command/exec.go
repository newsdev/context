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

const (
	ExecTemplateToken = `{}`
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

	// The key should have been saved as a binary, so no extra processing
	// should be needed.
	key, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if err := file.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}

	// No need to keep the file open.
	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Use the key to create a new crypter of the given type.
	c, err := crypter.NewCrypter(crypterType, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Create a new backend of the given type.
	b, err := backend.NewBackend(backendType, backendNamespace, backendProtocol, backendAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Get all of the values belonging to this group at once.
	ecryptedEnv, err := b.GetGroup(group)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// We want to use the current environment as a starting point and
	// overwrite values that are also specified in the encrypted environment.
	env := make(map[string]string)
	for _, variable := range os.Environ() {
		components := strings.Split(variable, `=`)
		env[components[0]] = components[1]
	}

	// In order to correctly pass templated arguments, the template must be
	// split along spaces. This doesn't take into account nested strings or
	// escape sequences, which could be problematic.
	//
	// TODO: Better template parsing method.
	templateSplit := strings.Split(template, ` `)
	templateArgs := make([]string, 0)

	for variable, encryptedValue := range ecryptedEnv {

		value, err := c.ValidateAndDecrypt(encryptedValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		env[variable] = string(value)

		for _, templateComponent := range templateSplit {
			templateArgs = append(templateArgs, strings.Replace(templateComponent, ExecTemplateToken, variable, -1))
		}
	}

	// Find the expanded path to the given executable.
	command, err := exec.LookPath(flagArgs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// Exec expects the environment to be specified as a slice rather than a
	// map. Flatten it by joining keys and values with `=`.
	commandEnv := make([]string, 0, len(env))
	for key, value := range env {
		commandEnv = append(commandEnv, fmt.Sprintf(`%s=%s`, key, value))
	}

	// Replace the template token in the given command argument slice with the
	// template arguments slice.
	commandArgs := make([]string, 0)
	for _, arg := range flagArgs.Args() {
		if arg == ExecTemplateToken {
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
