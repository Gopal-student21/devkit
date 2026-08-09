package contract

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// Issue is a single validation finding.
type Issue struct {
	Kind   string // schema | missing | type | extra | status | ref
	Path   string // JSON pointer-ish path
	Msg    string
	Strict bool // only reported in strict mode
}

// ValidateBytes validates a JSON body against a schema.
func ValidateBytes(spec *Spec, body []byte, schema *Schema, strict bool) ([]Issue, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return []Issue{{Kind: "schema", Path: "$", Msg: "invalid JSON: " + err.Error()}}, nil
	}
	return Validate(spec, data, schema, "$", strict, make(map[string]bool)), nil
}

// Validate walks data against schema, emitting Issues. Depth guards ref cycles.
func Validate(spec *Spec, data interface{}, schema *Schema, path string, strict bool, seen map[string]bool) []Issue {
	if schema == nil {
		return nil
	}
	if schema.Ref != "" {
		resolved, ok := spec.Resolve(schema.Ref)
		if !ok {
			return []Issue{{Kind: "ref", Path: path, Msg: "unresolved $ref " + schema.Ref}}
		}
		name := schema.SchemaName()
		if seen[name] {
			return nil
		}
		seen[name] = true
		defer delete(seen, name)
		return Validate(spec, data, resolved, path, strict, seen)
	}

	if data == nil {
		if schema.Nullable {
			return nil
		}
		return []Issue{{Kind: "schema", Path: path, Msg: "got null"}}
	}

	var issues []Issue
	switch schema.Type {
	case "object":
		issues = append(issues, validateObject(spec, data, schema, path, strict, seen)...)
	case "array":
		issues = append(issues, validateArray(spec, data, schema, path, strict, seen)...)
	case "string":
		s, ok := data.(string)
		if !ok {
			return []Issue{{Kind: "type", Path: path, Msg: fmt.Sprintf("expected string, got %T", data)}}
		}
		if len(schema.Enum) > 0 {
			ok := false
			for _, e := range schema.Enum {
				if fmt.Sprint(e) == s {
					ok = true
					break
				}
			}
			if !ok {
				issues = append(issues, Issue{Kind: "schema", Path: path, Msg: fmt.Sprintf("%q not in enum %v", s, schema.Enum)})
			}
		}
		if schema.Format == "email" && !strings.Contains(s, "@") {
			issues = append(issues, Issue{Kind: "schema", Path: path, Msg: "invalid email format"})
		}
	case "number", "integer":
		n, ok := asFloat(data)
		if !ok {
			return []Issue{{Kind: "type", Path: path, Msg: fmt.Sprintf("expected %s, got %T", schema.Type, data)}}
		}
		if schema.Type == "integer" && n != math.Trunc(n) {
			issues = append(issues, Issue{Kind: "type", Path: path, Msg: fmt.Sprintf("expected integer, got %v", n)})
		}
	case "boolean":
		if _, ok := data.(bool); !ok {
			return []Issue{{Kind: "type", Path: path, Msg: fmt.Sprintf("expected boolean, got %T", data)}}
		}
	case "null":
		if data != nil {
			return []Issue{{Kind: "type", Path: path, Msg: "expected null"}}
		}
	default:
		// no type constraint — pass
	}
	return issues
}

func validateObject(spec *Spec, data interface{}, schema *Schema, path string, strict bool, seen map[string]bool) []Issue {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return []Issue{{Kind: "type", Path: path, Msg: fmt.Sprintf("expected object, got %T", data)}}
	}
	var issues []Issue
	for _, req := range schema.Required {
		if _, ok := obj[req]; !ok {
			issues = append(issues, Issue{Kind: "missing", Path: path + "." + req, Msg: "required field missing"})
		}
	}
	for key, val := range obj {
		prop, defined := schema.Properties[key]
		if !defined {
			if strict {
				issues = append(issues, Issue{Kind: "extra", Path: path + "." + key, Msg: "field not in schema", Strict: true})
			}
			continue
		}
		issues = append(issues, Validate(spec, val, prop, path+"."+key, strict, seen)...)
	}
	return issues
}

func validateArray(spec *Spec, data interface{}, schema *Schema, path string, strict bool, seen map[string]bool) []Issue {
	arr, ok := data.([]interface{})
	if !ok {
		return []Issue{{Kind: "type", Path: path, Msg: fmt.Sprintf("expected array, got %T", data)}}
	}
	var issues []Issue
	for i, item := range arr {
		issues = append(issues, Validate(spec, item, schema.Items, fmt.Sprintf("%s[%d]", path, i), strict, seen)...)
	}
	return issues
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}
