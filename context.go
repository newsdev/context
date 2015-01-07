package main

import (
	"log"
	"os"

	"github.com/buth/context/command"
	"github.com/mitchellh/cli"
)

const (
	Version = "0.0.1"
)

func main() {
	c := cli.NewCLI("context", Version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"set": func() (cli.Command, error) {
			return &command.SetCommand{}, nil
		},
		"unset": func() (cli.Command, error) {
			return &command.UnsetCommand{}, nil
		},
		"key": func() (cli.Command, error) {
			return &command.KeyCommand{}, nil
		},
		"exec": func() (cli.Command, error) {
			return &command.ExecCommand{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
