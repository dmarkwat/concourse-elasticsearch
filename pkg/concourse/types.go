package concourse

import "github.com/dmarkwat/concourse-elasticsearch/pkg/es"

type SourceConfig struct {
	Addresses  []string `json:"addresses"`
	Index      string   `json:"index"`
	SortFields []string `json:"sort_fields"`
	Username   string   `json:"username,omitempty"`
	Password   string   `json:"password,omitempty"`
}

type InParams struct {
}

type OutParams struct {
	Document string `json:"document"`
	FieldMap map[string]es.PropertyMapping
}

type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Version struct {
	Id string `json:"id"`
}

type CheckRequest struct {
	Source  SourceConfig `json:"source"`
	Version *Version     `json:"version,omitempty"`
}

type CheckResponse []Version

type InRequest struct {
	Source  SourceConfig `json:"source"`
	Version Version      `json:"version"`
	Params  *InParams    `json:"params,omitempty"`
}

type InResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}

type OutRequest struct {
	Source SourceConfig `json:"source,omitempty"`
	Params *OutParams   `json:"params,omitempty"`
}

type OutResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}
