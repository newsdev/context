package backend

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

const (
	KeySeperator = "/"
)

type EtcdBackend struct {
	namespace, address string
	client             *etcd.Client
}

func NewEtcdBackend(namespace string, machines []string) *EtcdBackend {
	return &EtcdBackend{
		namespace: namespace,
		client:    etcd.NewClient(machines),
	}
}

func key(components ...string) string {
	return strings.Join(components, KeySeperator)
}

func (e *EtcdBackend) keyGroup(group string) string {
	return key(e.namespace, group)
}

func (e *EtcdBackend) keyVariable(group, variable string) string {
	return key(e.namespace, group, variable)
}

func (e *EtcdBackend) GetVariable(group, variable string) ([]byte, error) {
	response, err := e.client.Get(e.keyVariable(group, variable), false, false)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(response.Node.Value)
}

func (e *EtcdBackend) setVariable(group, variable string, value []byte, ttl uint64) error {
	encodedValue := base64.StdEncoding.EncodeToString(value)
	_, err := e.client.Set(e.keyVariable(group, variable), encodedValue, ttl)
	return err
}

func (e *EtcdBackend) SetVariable(group, variable string, value []byte) error {
	return e.setVariable(group, variable, value, 0)
}

func (e *EtcdBackend) RemoveVariable(group, variable string) error {
	_, err := e.client.Delete(e.keyVariable(group, variable), false)
	return err
}

func (e *EtcdBackend) GetGroup(group string) (map[string][]byte, error) {
	key := e.keyGroup(group)
	response, err := e.client.Get(key, false, true)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("/%s/", key)
	groupMap := make(map[string][]byte)
	for _, node := range response.Node.Nodes {
		value, err := base64.StdEncoding.DecodeString(node.Value)
		if err != nil {
			return nil, err
		}
		groupMap[strings.TrimPrefix(node.Key, prefix)] = value
	}

	return groupMap, nil
}

func (e *EtcdBackend) RemoveGroup(group string) error {
	_, err := e.client.Delete(e.keyGroup(group), true)
	return err
}
