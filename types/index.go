package types

import "time"

type StoreKind string

const (
	StoreKindFilesystem StoreKind = "filesystem"
	StoreKindVault      StoreKind = "vault"
)

type Index struct {
	Kind                 StoreKind         `yaml:"kind"` // vault / file
	ContextToPathMapping map[string]string `yaml:"contextToPathMapping"`
}

type IndexState struct {
	Kind           StoreKind `yaml:"kind"` // vault / file
	LastUpdateTime time.Time `yaml:"lastExecutionTime"`
}
