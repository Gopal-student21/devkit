package contract

import (
	"encoding/json"
	"testing"
)

const testSpec = `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /api/users:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/User"
  /api/users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
  /api/orders:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateOrderRequest"
      responses:
        "201":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Order"
components:
  schemas:
    User:
      type: object
      required: [id, email]
      properties:
        id: {type: string}
        email: {type: string, format: email}
        role: {type: string, enum: [admin, member]}
    Order:
      type: object
      required: [id, userId, total]
      properties:
        id: {type: string}
        userId: {type: string}
        total: {type: number}
    CreateOrderRequest:
      type: object
      required: [userId, total]
      properties:
        userId: {type: string}
        total: {type: number}
`

func mustSpec(t *testing.T) *Spec {
	t.Helper()
	s, err := ParseSpec([]byte(testSpec))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParseSpecEndpoints(t *testing.T) {
	s := mustSpec(t)
	eps := s.Endpoints()
	if len(eps) != 3 {
		t.Fatalf("expected 3 endpoints, got %d", len(eps))
	}
	methods := map[string]bool{}
	for _, e := range eps {
		methods[e.Method] = true
	}
	if !methods["get"] || !methods["post"] {
		t.Errorf("expected get+post, got %v", methods)
	}
}

func TestResolveRef(t *testing.T) {
	s := mustSpec(t)
	sch, ok := s.Resolve("#/components/schemas/User")
	if !ok || sch.Type != "object" {
		t.Error("failed to resolve User schema")
	}
	if _, ok := s.Resolve("#/components/schemas/Missing"); ok {
		t.Error("resolved a missing schema")
	}
}

func TestValidateValidUser(t *testing.T) {
	s := mustSpec(t)
	sch, _ := s.Resolve("#/components/schemas/User")
	body := `{"id":"u1","email":"a@b.com","role":"admin"}`
	issues, err := ValidateBytes(s, []byte(body), sch, false)
	if err != nil || len(issues) != 0 {
		t.Errorf("expected clean, got %+v %v", issues, err)
	}
}

func TestValidateMissingField(t *testing.T) {
	s := mustSpec(t)
	sch, _ := s.Resolve("#/components/schemas/User")
	issues, _ := ValidateBytes(s, []byte(`{"id":"u1"}`), sch, false)
	found := false
	for _, i := range issues {
		if i.Kind == "missing" && i.Path == "$.email" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing-field issue, got %+v", issues)
	}
}

func TestValidateWrongType(t *testing.T) {
	s := mustSpec(t)
	sch, _ := s.Resolve("#/components/schemas/User")
	issues, _ := ValidateBytes(s, []byte(`{"id":"u1","email":42}`), sch, false)
	found := false
	for _, i := range issues {
		if i.Kind == "type" && i.Path == "$.email" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected type issue on email, got %+v", issues)
	}
}

func TestValidateEnumViolation(t *testing.T) {
	s := mustSpec(t)
	sch, _ := s.Resolve("#/components/schemas/User")
	issues, _ := ValidateBytes(s, []byte(`{"id":"u1","email":"a@b.com","role":"root"}`), sch, false)
	found := false
	for _, i := range issues {
		if i.Kind == "schema" && i.Path == "$.role" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected enum violation, got %+v", issues)
	}
}

func TestValidateExtraFieldStrict(t *testing.T) {
	s := mustSpec(t)
	sch, _ := s.Resolve("#/components/schemas/User")
	body := []byte(`{"id":"u1","email":"a@b.com","hack":"x"}`)
	// non-strict: extra allowed
	if issues, _ := ValidateBytes(s, body, sch, false); len(issues) != 0 {
		t.Errorf("non-strict should allow extras, got %+v", issues)
	}
	// strict: extra flagged
	issues, _ := ValidateBytes(s, body, sch, true)
	found := false
	for _, i := range issues {
		if i.Kind == "extra" && i.Path == "$.hack" {
			found = true
		}
	}
	if !found {
		t.Errorf("strict should flag extra field, got %+v", issues)
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	s := mustSpec(t)
	sch, _ := s.Resolve("#/components/schemas/User")
	issues, _ := ValidateBytes(s, []byte("{nope"), sch, false)
	if len(issues) == 0 {
		t.Error("expected invalid-JSON issue")
	}
}

func TestSeedReferentialIntegrity(t *testing.T) {
	s := mustSpec(t)
	seeder := NewSeeder(s, SeedOptions{Count: 5, Seed: 42})
	records := seeder.Generate()

	if len(records["Order"]) != 5 || len(records["User"]) != 5 {
		t.Fatalf("expected 5 records each, got users=%d orders=%d", len(records["User"]), len(records["Order"]))
	}

	// every Order.userId must exist in the generated User set
	userIDs := map[string]bool{}
	for _, u := range records["User"] {
		id, _ := u["id"].(string)
		userIDs[id] = true
	}
	for i, o := range records["Order"] {
		uid, ok := o["userId"].(string)
		if !ok {
			t.Fatalf("order[%d].userId missing/non-string: %v", i, o["userId"])
		}
		if !userIDs[uid] {
			t.Errorf("order[%d] references orphaned userId %q", i, uid)
		}
	}
}

func TestSeedDeterministic(t *testing.T) {
	s := mustSpec(t)
	a := NewSeeder(s, SeedOptions{Count: 3, Seed: 99}).Generate()
	b := NewSeeder(s, SeedOptions{Count: 3, Seed: 99}).Generate()
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Error("same seed should produce identical output")
	}
}

func TestSeedSQL(t *testing.T) {
	s := mustSpec(t)
	seeder := NewSeeder(s, SeedOptions{Count: 2, Seed: 1})
	seeder.Generate()
	sql, err := seeder.ToSQL()
	if err != nil {
		t.Fatal(err)
	}
	if len(sql) == 0 {
		t.Error("expected SQL output")
	}
}
