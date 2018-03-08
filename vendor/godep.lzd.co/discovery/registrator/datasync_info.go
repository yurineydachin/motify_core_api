package registrator

import (
	"fmt"
	"strings"

	"godep.lzd.co/discovery/provider"
)

const (
	keyPartSeparator = "/"
)

// ExportedEntity is a data struct for DataSync API.
type ExportedEntity struct {
	Name     string
	Endpoint string
}

// NewExportedEntityFromKey returns new ExportedEntity from key string
func NewExportedEntityFromKey(key string) (ExportedEntity, error) {
	parts := strings.Split(key, "/")
	if len(parts) != 4 {
		return ExportedEntity{}, fmt.Errorf("can't parse key '%s': parts number missmatch, 4 parts expected", key)
	}
	if parts[1] != provider.NamespaceExportedEntities {
		return ExportedEntity{}, fmt.Errorf("can't parse key '%s': invalid namespace, '%s' expected", key, provider.NamespaceExportedEntities)
	}

	return ExportedEntity{
		Name:     parts[2],
		Endpoint: parts[3],
	}, nil
}

// kv converts ExportedEntity into provider.KV struct
func (i ExportedEntity) kv() provider.KV {
	return provider.KV{
		RawKey: fmt.Sprintf("/%s/%s/%s", provider.NamespaceExportedEntities, i.Name, i.Endpoint),
	}
}

type exportedNames []string

func (e exportedNames) newKVs(endpoint string) []provider.KV {
	if endpoint == "" {
		return nil
	}

	kvs := make([]provider.KV, 0, len(e))
	for _, name := range e {
		kvs = append(kvs, ExportedEntity{Name: name, Endpoint: endpoint}.kv())
	}
	return kvs
}

func (e exportedNames) validate() error {
	for _, name := range e {
		if strings.Contains(name, keyPartSeparator) {
			return fmt.Errorf("invalid ExportedEntity name: '%s' contains forbidden '/' char", name)
		}
	}
	return nil
}
