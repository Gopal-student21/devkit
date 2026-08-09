package contract

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// SeedOptions controls mock-data generation.
type SeedOptions struct {
	Count int   // records per schema (default 5)
	Seed  int64 // PRNG seed for deterministic output
	Link  int   // FK lookup depth for cyclic schemas
}

// Seeder generates referentially-consistent mock data from a Spec.
// Foreign-key fields point at records that actually exist (generated first),
// unlike naive faker output which produces orphaned references.
type Seeder struct {
	spec  *Spec
	rng   *rand.Rand
	count int
	link  int
	// records[schemaName] = generated records
	records map[string][]map[string]interface{}
	// edges["<schema>:<prop>"] = target schema name (FK detection)
	edges map[string]string
}

// NewSeeder builds a Seeder from a Spec.
func NewSeeder(spec *Spec, opts SeedOptions) *Seeder {
	if opts.Count <= 0 {
		opts.Count = 5
	}
	if opts.Seed == 0 {
		opts.Seed = 42
	}
	if opts.Link <= 0 {
		opts.Link = 3
	}
	return &Seeder{
		spec:    spec,
		rng:     rand.New(rand.NewSource(opts.Seed)),
		count:   opts.Count,
		link:    opts.Link,
		records: map[string][]map[string]interface{}{},
		edges:   map[string]string{},
	}
}

// detectFKs finds relationships: $ref fields (explicit) and <name>Id / <name>_id
// naming conventions pointing at an existing schema (implicit).
func (s *Seeder) detectFKs() {
	for _, schemaName := range sortedKeys(s.spec.Components.Schemas) {
		schema := s.spec.Components.Schemas[schemaName]
		for _, propName := range sortedKeys(schema.Properties) {
			prop := schema.Properties[propName]
			if target, ok := s.refTarget(prop); ok {
				s.edges[schemaName+":"+propName] = target
			}
		}
	}
}

// refTarget returns the schema a $ref property points to.
func (s *Seeder) refTarget(prop *Schema) (string, bool) {
	if prop == nil || prop.Ref == "" {
		return "", false
	}
	t := prop.SchemaName()
	if _, ok := s.spec.Components.Schemas[t]; ok {
		return t, true
	}
	return "", false
}

// recordsFor returns (and lazily generates) the records for a schema.
func (s *Seeder) recordsFor(name string, depth int) []map[string]interface{} {
	if recs, ok := s.records[name]; ok {
		return recs
	}
	if depth > s.link {
		s.records[name] = []map[string]interface{}{}
		return s.records[name]
	}
	schema, ok := s.spec.Components.Schemas[name]
	if !ok {
		s.records[name] = nil
		return nil
	}
	recs := make([]map[string]interface{}, s.count)
	for i := 0; i < s.count; i++ {
		recs[i] = s.generateObject(schema, name, i, depth)
	}
	s.records[name] = recs
	return recs
}

// generateObject builds one record for a schema.
func (s *Seeder) generateObject(schema *Schema, name string, idx, depth int) map[string]interface{} {
	obj := map[string]interface{}{}
	for _, propName := range sortedKeys(schema.Properties) {
		prop := schema.Properties[propName]
		obj[propName] = s.generateValue(prop, propName, name, idx, depth)
	}
	return obj
}

// generateValue produces a value for a property, honoring FK references.
func (s *Seeder) generateValue(prop *Schema, propName, parent string, idx, depth int) interface{} {
	if prop == nil {
		return nil
	}
	if prop.Ref != "" {
		target := prop.SchemaName()
		if _, ok := s.spec.Components.Schemas[target]; ok {
			recs := s.recordsFor(target, depth+1)
			if len(recs) == 0 {
				return nil
			}
			return s.fkValue(recs)
		}
		return map[string]interface{}{}
	}

	// implicit FK via naming convention (e.g. userId, order_id)
	if prop.Type == "string" && propName != "id" && strings.HasSuffix(strings.ToLower(propName), "id") {
		base := strings.TrimSuffix(propName, "Id")
		base = strings.TrimSuffix(base, "_id")
		for _, schemaName := range sortedKeys(s.spec.Components.Schemas) {
			if strings.EqualFold(schemaName, base) || strings.EqualFold(schemaName, base+"s") {
				recs := s.recordsFor(schemaName, depth+1)
				if len(recs) > 0 {
					return s.fkValue(recs)
				}
			}
		}
	}

	switch prop.Type {
	case "string":
		return s.genString(prop, propName, parent, idx)
	case "integer":
		return idx + 1
	case "number":
		return float64(idx+1)*1.5 + float64(s.rng.Intn(100))/100.0
	case "boolean":
		return s.rng.Intn(2) == 0
	case "array":
		n := 1 + s.rng.Intn(3)
		arr := make([]interface{}, 0, n)
		for i := 0; i < n; i++ {
			arr = append(arr, s.generateValue(prop.Items, propName, parent, idx+i, depth))
		}
		return arr
	case "object":
		if len(prop.Properties) == 0 {
			return map[string]interface{}{}
		}
		return s.generateObject(prop, propName, idx, depth)
	default:
		return nil
	}
}

// fkValue picks a stable, existing primary key from a generated record set.
func (s *Seeder) fkValue(recs []map[string]interface{}) interface{} {
	idx := s.rng.Intn(len(recs))
	if pk, ok := recs[idx]["id"]; ok {
		return pk
	}
	return recs[idx]
}

// genString produces a deterministic, format-aware string.
func (s *Seeder) genString(prop *Schema, propName, parent string, idx int) string {
	if len(prop.Enum) > 0 {
		return fmt.Sprint(prop.Enum[idx%len(prop.Enum)])
	}
	switch prop.Format {
	case "email":
		return fmt.Sprintf("%s%d@example.com", strings.ToLower(propName), idx+1)
	case "date-time":
		return time.Unix(int64(idx)*86400+1, 0).UTC().Format(time.RFC3339)
	case "uuid":
		sum := md5.Sum([]byte(fmt.Sprintf("%s-%d", propName, idx)))
		return fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
	case "url", "uri":
		return fmt.Sprintf("https://example.com/%s/%d", strings.ToLower(propName), idx+1)
	}
	if strings.EqualFold(propName, "id") {
		return fmt.Sprintf("%s-%04d", strings.ToLower(parent), idx+1)
	}
	return fmt.Sprintf("%s-%d", strings.ToLower(propName), idx+1)
}

// Generate produces all schemas' records, dependency-first.
func (s *Seeder) Generate() map[string][]map[string]interface{} {
	s.detectFKs()
	for _, name := range s.topologicalOrder() {
		s.recordsFor(name, 0)
	}
	return s.records
}

// topologicalOrder sorts schemas so referenced schemas generate first.
func (s *Seeder) topologicalOrder() []string {
	names := sortedKeys(s.spec.Components.Schemas)
	state := map[string]int{} // 0 unvisited, 1 visiting, 2 done
	var out []string
	var visit func(n string)
	visit = func(n string) {
		if state[n] != 0 {
			return
		}
		state[n] = 1
		for _, other := range names {
			for _, prop := range sortedKeys(s.spec.Components.Schemas[other].Properties) {
				if s.edges[other+":"+prop] == n {
					visit(other)
				}
			}
		}
		state[n] = 2
		out = append(out, n)
	}
	for _, n := range names {
		visit(n)
	}
	return out
}

// ToJSON renders the full generated dataset.
func (s *Seeder) ToJSON() (string, error) {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToSQL renders INSERT statements (best-effort table naming).
func (s *Seeder) ToSQL() (string, error) {
	var b strings.Builder
	for _, name := range sortedKeys(s.records) {
		recs := s.records[name]
		if len(recs) == 0 {
			continue
		}
		table := strings.ToLower(name)
		cols := sortedKeys(recs[0])
		b.WriteString(fmt.Sprintf("INSERT INTO %q (%s) VALUES\n", table, strings.Join(cols, ", ")))
		for i, r := range recs {
			var vals []string
			for _, c := range cols {
				vals = append(vals, sqlValue(r[c]))
			}
			end := ","
			if i == len(recs)-1 {
				end = ";"
			}
			b.WriteString(fmt.Sprintf("  (%s)%s\n", strings.Join(vals, ", "), end))
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

func sqlValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(t, "'", "''") + "'"
	case nil:
		return "NULL"
	case []interface{}, map[string]interface{}:
		enc, err := json.Marshal(t)
		if err != nil {
			return "NULL"
		}
		return "'" + strings.ReplaceAll(string(enc), "'", "''") + "'"
	default:
		return fmt.Sprint(v)
	}
}
