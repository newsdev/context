package command

import (
	"os"
	"testing"

	"github.com/nytinteractive/context/backend"
	"github.com/nytinteractive/context/crypter/std"
)

func TestSetCommand(t *testing.T) {

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

	exitCode := s.Run([]string{"set", "A"}, map[string]string{"GROUP": "testGroup", "A": "value"}, os.Stdout)
	if exitCode != 0 {
		t.Fatalf("set command exited with non-zero status: %d", exitCode)
	}

	groupEnv, err := b.GetGroup("testGroup")
	if err != nil {
		t.Fatal(err)
	}

	value, err := c.ValidateAndDecrypt(groupEnv["A"])
	if err != nil {
		t.Fatal(err)
	}

	if valueString := string(value); valueString != "value" {
		t.Errorf("retrieved value did not match: expected \"%s\", go \"%s\"", "value", valueString)
	}
}
