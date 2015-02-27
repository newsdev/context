package backend

import (
	"fmt"
	"strings"
)

type Backend interface {
	GetVariable(group, variable string) ([]byte, error)
	SetVariable(group, variable string, value []byte) error
	RemoveVariable(group, variable string) error
	GetGroup(group string) (map[string][]byte, error)
	RemoveGroup(group string) error
}

func NewBackend(kind, namespace, address string) (Backend, error) {

	// Select a backend based on kind.
	switch kind {
	case "etcd":
		backend := NewEtcdBackend(namespace, strings.Split(address, ","))
		return backend, nil
	case "redis":
		backend := NewRedisBackend(namespace, address)
		return backend, nil
	}

	// Assuming no backend is implemented for kind.
	return nil, NoBackendError{kind}
}

type NoBackendError struct {
	Kind string
}

func (e NoBackendError) Error() string {
	return fmt.Sprintf("backend: backend \"%s\" has not been implemented", e.Kind)
}
