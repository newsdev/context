package command

import (
	"os"
	"testing"

	"github.com/nytinteractive/context/backend"
	"github.com/nytinteractive/context/crypter/std"
)

func TestUnsetCommand(t *testing.T) {

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

	u := &UnsetCommand{
		Crypter: c,
		Backend: b,
	}

	if exitCode := u.Run([]string{"unset", "A"}, map[string]string{"GROUP": "testGroup"}, os.Stdout); exitCode != 0 {
		t.Fatalf("unset command exited with non-zero status: %d", exitCode)
	}

	groupEnv, err := b.GetGroup("testGroup")
	if err != nil {
		t.Fatal(err)
	}

	if groupEnv["A"] != nil {
		t.Errorf("retrieved value was not nil: expected nil, go \"%s\"", string(groupEnv["A"]))
	}
}
