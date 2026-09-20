package compiler

import (
	"strings"
	"testing"

	"github.com/peacewalker122/mapper/idgen"
)

const testSchema = `version: 1
model:
  name: subscriber
  fields:
    - name: msisdn
      type: string
      required: true
    - name: status
      type: string
      required: true
    - name: last_tx
      type: datetime
`

func compileForTest(t *testing.T, source, lock []byte) *CompileResult {
	t.Helper()
	result, err := (&Compiler{IDGenerator: idgen.NewSequentialIDGenerator(1001)}).Compile(CompileRequest{
		Source:   source,
		LockFile: lock,
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	return result
}

func TestParseAndValidate(t *testing.T) {
	parsed, err := ParseAndValidate([]byte(testSchema))
	if err != nil {
		t.Fatalf("ParseAndValidate: %v", err)
	}
	if parsed.Model.Name != "subscriber" || len(parsed.Model.Fields) != 3 {
		t.Fatalf("parsed schema = %#v", parsed)
	}
}

func TestCompilePreservesIDsWithLock(t *testing.T) {
	first := compileForTest(t, []byte(testSchema), nil)
	second := compileForTest(t, []byte(testSchema), first.LockFile)
	if first.Schema.Model.ID != second.Schema.Model.ID {
		t.Fatalf("schema ID changed: %d -> %d", first.Schema.Model.ID, second.Schema.Model.ID)
	}
	for index, field := range first.Schema.Model.Fields {
		if field.ID != second.Schema.Model.Fields[index].ID {
			t.Fatalf("field %q ID changed: %d -> %d", field.Name, field.ID, second.Schema.Model.Fields[index].ID)
		}
	}
}

func TestCompileRemovalAndReaddPreservesFieldID(t *testing.T) {
	first := compileForTest(t, []byte(testSchema), nil)
	statusID := first.Schema.Model.Fields[1].ID
	reduced := []byte(`version: 1
model:
  name: subscriber
  fields:
    - name: msisdn
      type: string
      required: true
    - name: last_tx
      type: datetime
`)
	removed := compileForTest(t, reduced, first.LockFile)
	lock, err := ParseLock(removed.LockFile)
	if err != nil {
		t.Fatalf("ParseLock: %v", err)
	}
	if lock.Fields["status"].Status != FieldRemoved {
		t.Fatalf("status state = %q", lock.Fields["status"].Status)
	}
	readded := compileForTest(t, []byte(testSchema), removed.LockFile)
	if got := readded.Schema.Model.Fields[1].ID; got != statusID {
		t.Fatalf("re-added status ID = %d, want %d", got, statusID)
	}
	if lock, err = ParseLock(readded.LockFile); err != nil || lock.Fields["status"].Status != FieldActive {
		t.Fatalf("re-added status lock state = %#v, err = %v", lock.Fields["status"], err)
	}
}

func TestValidationRejectsUnsupportedSchema(t *testing.T) {
	_, err := ParseAndValidate([]byte(strings.Replace(testSchema, "datetime", "strng", 1)))
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("error = %v", err)
	}
}
