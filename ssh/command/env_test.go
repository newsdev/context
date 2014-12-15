package command

import (
	"bytes"
	"os"
	"testing"

	"github.com/nytinteractive/context/backend"
	"github.com/nytinteractive/context/crypter/std"
)

func TestEnvCommand(t *testing.T) {

	_, _, c, err := std.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	b, err := backend.NewBackend("redis", "test", "tcp", ":6379")
	if err != nil {
		t.Fatal(err)
	}

	s := &SetCommand{
		Crypter: c,
		Backend: b,
	}

	if exitCode := s.Run([]string{"set", "A"}, map[string]string{"GROUP": "testGroup", "A": "value"}, os.Stdout); exitCode != 0 {
		t.Fatalf("set command exited with non-zero status: %d", exitCode)
	}

	e := &EnvCommand{
		Crypter: c,
		Backend: b,
	}

	buf := bytes.NewBuffer([]byte{})
	if exitCode := e.Run([]string{"env"}, map[string]string{"GROUP": "testGroup"}, buf); exitCode != 0 {
		t.Fatalf("env command exited with non-zero status: %d", exitCode)
	}

	if envValue := buf.String(); envValue != "A=value\n" {
		t.Errorf("retrieved value did not match: expected \"%s\", go \"%s\"", "A=value\n", envValue)
	}
}
