package backend

import (
	"fmt"
	"github.com/buth/stocker/backend/etcd"
	"github.com/buth/stocker/backend/redis"
	"time"
)

type Backend interface {
	GetVariable(group, variable string) ([]byte, error)
	SetVariable(group, variable string, value []byte) error
	SetVariableTTL(group, variable string, value []byte, ttl time.Duration) error
	RemoveVariable(group, variable string) error
	GetGroup(group string) (map[string][]byte, error)
	RemoveGroup(group string) error
}

func NewBackend(kind, namespace, protocol, address string) (Backend, error) {

	// Select a backend based on kind.
	switch kind {
	case "etcd":
		backend := etcd.New(namespace, protocol, address)
		return backend, nil
	case "redis":
		backend := redis.New(namespace, protocol, address)
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
