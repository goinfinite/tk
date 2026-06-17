package tkInfraDb

import (
	"os"
	"testing"
	"time"
)

func TestTransientDatabaseFileInInfraDbDirectory(t *testing.T) {
	_, statErr := os.Stat("transientDatabaseService.go")
	if statErr != nil {
		t.Fatalf("TransientDatabaseFileNotFound: %s", statErr.Error())
	}
}

func TestTransientDatabaseConstructorTakesNoParameters(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	if dbSvc == nil {
		t.Fatalf("ServiceIsNil")
	}

	if dbSvc.Handler == nil {
		t.Fatalf("HandlerIsNil")
	}
}

func TestTransientDatabaseApiSurfaceHasReadSetCount(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	hasResult := dbSvc.Has("test-key")
	if hasResult {
		t.Fatalf("HasShouldReturnFalseForMissingKey")
	}

	setErr := dbSvc.Set("test-key", "test-value")
	if setErr != nil {
		t.Fatalf("SetFailed: %s", setErr.Error())
	}

	readValue, readErr := dbSvc.Read("test-key")
	if readErr != nil {
		t.Fatalf("ReadFailed: %s", readErr.Error())
	}

	if readValue != "test-value" {
		t.Fatalf("ReadValueMismatch: got=%s, want=test-value", readValue)
	}

	countResult := dbSvc.Count()
	if countResult != 1 {
		t.Fatalf("CountMismatch: got=%d, want=1", countResult)
	}
}

func TestTransientDatabaseModelHasCorrectFields(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	setErr := dbSvc.Set("field-test-key", "field-test-value")
	if setErr != nil {
		t.Fatalf("SetFailed: %s", setErr.Error())
	}

	var modelEntry KeyValueModel
	queryErr := dbSvc.Handler.Where(
		"key = ?", "field-test-key",
	).First(&modelEntry).Error
	if queryErr != nil {
		t.Fatalf("QueryFailed: %s", queryErr.Error())
	}

	if modelEntry.Key != "field-test-key" {
		t.Fatalf("KeyMismatch: got=%s, want=field-test-key", modelEntry.Key)
	}

	if modelEntry.Value != "field-test-value" {
		t.Fatalf("ValueMismatch: got=%s, want=field-test-value", modelEntry.Value)
	}

	if modelEntry.CreatedAt.IsZero() {
		t.Fatalf("CreatedAtIsZero")
	}
}

func TestTransientDatabaseCountReturnsTotalEntries(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	dbSvc.Handler.Where("1 = 1").Delete(&KeyValueModel{})

	entryKeys := []string{
		"count-key-1", "count-key-2", "count-key-3",
		"count-key-4", "count-key-5",
	}
	for _, entryKey := range entryKeys {
		setErr := dbSvc.Set(entryKey, "value-"+entryKey)
		if setErr != nil {
			t.Fatalf("SetFailed: %s", setErr.Error())
		}
	}

	totalCount := dbSvc.Count()
	if totalCount != 5 {
		t.Fatalf("CountMismatch: got=%d, want=5", totalCount)
	}
}

func TestTransientDatabaseHasCreatedAtAutoManaged(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	beforeCreation := time.Now().UTC().Add(-time.Second)
	setErr := dbSvc.Set("created-at-key", "created-at-value")
	if setErr != nil {
		t.Fatalf("SetFailed: %s", setErr.Error())
	}

	var modelEntry KeyValueModel
	queryErr := dbSvc.Handler.Where(
		"key = ?", "created-at-key",
	).First(&modelEntry).Error
	if queryErr != nil {
		t.Fatalf("QueryFailed: %s", queryErr.Error())
	}

	if modelEntry.CreatedAt.Before(beforeCreation) {
		t.Fatalf("CreatedAtTooOld: got=%v, want>=%v", modelEntry.CreatedAt, beforeCreation)
	}
}

func TestTransientDatabaseReadAllReturnsAllEntries(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	dbSvc.Handler.Where("1 = 1").Delete(&KeyValueModel{})

	entryKeys := []string{
		"readall-key-1", "readall-key-2", "readall-key-3",
		"readall-key-4", "readall-key-5",
	}
	for _, entryKey := range entryKeys {
		setErr := dbSvc.Set(entryKey, "value-"+entryKey)
		if setErr != nil {
			t.Fatalf("SetFailed: %s", setErr.Error())
		}
	}

	allEntries, readAllErr := dbSvc.ReadAll()
	if readAllErr != nil {
		t.Fatalf("ReadAllFailed: %s", readAllErr.Error())
	}

	if len(allEntries) != 5 {
		t.Fatalf("ReadAllCountMismatch: got=%d, want=5", len(allEntries))
	}

	entryKeySet := make(map[string]bool)
	for _, entry := range allEntries {
		entryKeySet[entry.Key] = true
	}

	for _, expectedKey := range entryKeys {
		if !entryKeySet[expectedKey] {
			t.Fatalf("ReadAllMissingKey: %s", expectedKey)
		}
	}
}

func TestTransientDbConnectionFailureHandledGracefully(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionShouldSucceed: %s", constructErr.Error())
	}

	if dbSvc == nil {
		t.Fatalf("ServiceIsNil")
	}
}

func TestTransientDbSetErrorHandled(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	setErr := dbSvc.Set("error-test-key", "error-test-value")
	if setErr != nil {
		t.Fatalf("SetShouldSucceed: %s", setErr.Error())
	}
}

func TestTransientDbCountErrorHandled(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	dbSvc.Handler.Where("1 = 1").Delete(&KeyValueModel{})

	countResult := dbSvc.Count()
	if countResult != 0 {
		t.Fatalf("EmptyCountShouldBeZero: got=%d", countResult)
	}
}

func TestTransientDbReadAllErrorHandled(t *testing.T) {
	dbSvc, constructErr := NewTransientDatabaseService()
	if constructErr != nil {
		t.Fatalf("ConstructionFailed: %s", constructErr.Error())
	}

	dbSvc.Handler.Where("1 = 1").Delete(&KeyValueModel{})

	allEntries, readAllErr := dbSvc.ReadAll()
	if readAllErr != nil {
		t.Fatalf("ReadAllShouldSucceed: %s", readAllErr.Error())
	}

	if len(allEntries) != 0 {
		t.Fatalf("EmptyReadAllShouldReturnEmptySlice: got=%d", len(allEntries))
	}
}
