package contract

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec is a typed view of an OpenAPI 3.0 document (api.yaml).
type Spec struct {
	OpenAPI    string                 `yaml:"openapi"`
	Info       map[string]interface{} `yaml:"info"`
	Paths      map[string]*PathItem   `yaml:"paths"`
	Components struct {
		Schemas map[string]*Schema `yaml:"schemas"`
	} `yaml:"components"`
}

// PathItem holds the operations for a single path.
type PathItem struct {
	Operations map[string]*Operation
}

// Operation describes one method on a path.
type Operation struct {
	Summary     string
	Parameters  []*Parameter
	RequestBody *RequestBody
	Responses   map[string]*Response
}

// Parameter is a path/query/header parameter.
type Parameter struct {
	Name     string
	In       string
	Required bool
	Schema   *Schema
}

// RequestBody is the operation request body.
type RequestBody struct {
	Required bool
	Content  map[string]*MediaType
}

// Response is an operation response.
type Response struct {
	Description string
	Content     map[string]*MediaType
}

// MediaType maps content-type to a schema.
type MediaType struct {
	Schema *Schema
}

// Schema is an OpenAPI schema node.
type Schema struct {
	Type       string
	Format     string
	Ref        string             `yaml:"$ref"`
	Enum       []interface{}      `yaml:"enum"`
	Required   []string           `yaml:"required"`
	Properties map[string]*Schema `yaml:"properties"`
	Items      *Schema            `yaml:"items"`
	Nullable   bool               `yaml:"nullable"`
}

// UnmarshalYAML handles the polymorphic PathItem (map of methods).
func (p *PathItem) UnmarshalYAML(node *yaml.Node) error {
	raw := map[string]*Operation{}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	p.Operations = map[string]*Operation{}
	for k, v := range raw {
		m := strings.ToLower(k)
		switch m {
		case "get", "post", "put", "delete", "patch", "head", "options":
			p.Operations[m] = v
		}
	}
	return nil
}

// UnmarshalYAML handles the polymorphic Parameter (schema or content).
func (p *Parameter) UnmarshalYAML(node *yaml.Node) error {
	type alias Parameter
	var a struct {
		Name     string      `yaml:"name"`
		In       string      `yaml:"in"`
		Required bool        `yaml:"required"`
		Schema   *Schema     `yaml:"schema"`
		Content  interface{} `yaml:"content"`
	}
	if err := node.Decode(&a); err != nil {
		return err
	}
	p.Name, p.In, p.Required, p.Schema = a.Name, a.In, a.Required, a.Schema
	return nil
}

// LoadSpec reads and parses an OpenAPI file.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return ParseSpec(data)
}

// ParseSpec parses OpenAPI YAML bytes into a Spec.
func ParseSpec(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	if spec.OpenAPI == "" {
		return nil, fmt.Errorf("not an OpenAPI document (missing 'openapi' version)")
	}
	return &spec, nil
}

// Resolve follows a $ref like #/components/schemas/User to its Schema.
func (s *Spec) Resolve(ref string) (*Schema, bool) {
	trimmed := strings.TrimPrefix(ref, "#/components/schemas/")
	if trimmed == ref {
		return nil, false
	}
	sch, ok := s.Components.Schemas[trimmed]
	return sch, ok
}

// SchemaName returns the referenced schema name or the ref tail.
func (s *Schema) SchemaName() string {
	if s.Ref != "" {
		return s.Ref[strings.LastIndex(s.Ref, "/")+1:]
	}
	return ""
}

// Endpoints lists all path+method combinations with a JSON response.
func (s *Spec) Endpoints() []Endpoint {
	var eps []Endpoint
	for path, item := range s.Paths {
		for method, op := range item.Operations {
			eps = append(eps, Endpoint{Path: path, Method: method, Operation: op})
		}
	}
	return eps
}

// Endpoint is one path+method pair.
type Endpoint struct {
	Path      string
	Method    string
	Operation *Operation
}

// ResponseSchema returns the JSON schema for a given status (defaulting to 200).
func (e Endpoint) ResponseSchema(status string) (*Schema, bool) {
	resp, ok := e.Operation.Responses[status]
	if !ok {
		resp, ok = e.Operation.Responses["200"]
		if !ok {
			return nil, false
		}
	}
	mt, ok := resp.Content["application/json"]
	if !ok {
		return nil, false
	}
	return mt.Schema, mt.Schema != nil
}
