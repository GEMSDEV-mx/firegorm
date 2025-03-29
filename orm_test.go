// orm_test.go
package firegorm

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestMain initializes the logger before any tests are run.
func TestMain(m *testing.M) {
	InitializeLogger()
	os.Exit(m.Run())
}

// --- Test for ValidateStruct ---

// TestStruct is used for testing ValidateStruct.
type TestStruct struct {
	Name string `validate:"required"`
	Age  int
}

func TestValidateStruct_Success(t *testing.T) {
	s := TestStruct{Name: "John", Age: 30}
	if err := ValidateStruct(s); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestValidateStruct_Failure(t *testing.T) {
	s := TestStruct{Name: "", Age: 30}
	if err := ValidateStruct(s); err == nil {
		t.Error("expected error for missing required field, got nil")
	}
}

// --- Test for generateUUID ---

func TestGenerateUUID(t *testing.T) {
	id := generateUUID()
	if id == "" {
		t.Error("expected non-empty UUID")
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Errorf("generated UUID is not valid: %v", err)
	}
}

// --- Test for updatesToFirestoreUpdates ---
// This function converts a map to a slice of Firestore update structs.
func TestUpdatesToFirestoreUpdates(t *testing.T) {
	updates := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}
	fsUpdates := updatesToFirestoreUpdates(updates)
	if len(fsUpdates) != len(updates) {
		t.Errorf("expected %d updates, got %d", len(updates), len(fsUpdates))
	}
	// Verify that each update in the result matches the input.
	for _, upd := range fsUpdates {
		if v, ok := updates[upd.Path]; !ok {
			t.Errorf("unexpected update path: %s", upd.Path)
		} else if !reflect.DeepEqual(v, upd.Value) {
			t.Errorf("for key %s, expected value %v, got %v", upd.Path, v, upd.Value)
		}
	}
}



func TestParseFilter_Date(t *testing.T) {
	field, op, newVal, err := parseFilter("event_date__gte", "2025-04-01")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if field != "event_date" {
		t.Errorf("expected field 'event_date', got '%s'", field)
	}
	if op != ">=" {
		t.Errorf("expected operator '>=', got '%s'", op)
	}
	expectedTime, _ := time.Parse("2006-01-02", "2025-04-01")
	if !reflect.DeepEqual(newVal, expectedTime) {
		t.Errorf("expected value %v, got %v", expectedTime, newVal)
	}
}

func TestParseFilter_NonDate(t *testing.T) {
	field, op, newVal, err := parseFilter("status", "active")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if field != "status" {
		t.Errorf("expected field 'status', got '%s'", field)
	}
	if op != "==" {
		t.Errorf("expected operator '==', got '%s'", op)
	}
	if newVal != "active" {
		t.Errorf("expected value 'active', got '%v'", newVal)
	}
}

func TestParseFilter_InvalidDate(t *testing.T) {
	_, _, _, err := parseFilter("birthdate__lt", "invalid-date")
	if err == nil {
		t.Error("expected error for invalid date format, got nil")
	}
}

// --- Test for validateUpdateFields ---
// We define a dummy model and register it so that we can test the update validation.
type DummyModel struct {
	BaseModel
	// Field1 is required.
	Field1 string `firestore:"field1" json:"field1" validate:"required"`
	// DateField is an example date field.
	DateField time.Time `firestore:"date_field" json:"date_field"`
}

func TestValidateUpdateFields_Success(t *testing.T) {
	// Clear registry to ensure isolation
	modelRegistry = make(map[string]ModelInfo)
	
	dummy := DummyModel{}
	// Register the dummy model.
	_, err := RegisterModel(&dummy, "dummy")
	if err != nil {
		t.Fatalf("failed to register dummy model: %v", err)
	}
	// Ensure the model's BaseModel fields are set.
	dummy.SetCollectionName("dummy")
	dummy.SetModelName("DummyModel")

	updates := map[string]interface{}{
		"field1": "new value",
	}
	if err := validateUpdateFields(updates, &dummy.BaseModel); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestValidateUpdateFields_Failure(t *testing.T) {
	// Clear registry to ensure isolation
	modelRegistry = make(map[string]ModelInfo)
	
	dummy := DummyModel{}
	// Register the dummy model.
	_, err := RegisterModel(&dummy, "dummy")
	if err != nil {
		t.Fatalf("failed to register dummy model: %v", err)
	}
	dummy.SetCollectionName("dummy")
	dummy.SetModelName("DummyModel")

	// Use a field that does not exist in DummyModel.
	updates := map[string]interface{}{
		"nonexistent_field": "value",
	}
	if err := validateUpdateFields(updates, &dummy.BaseModel); err == nil {
		t.Error("expected error for invalid update field, got nil")
	}
}
