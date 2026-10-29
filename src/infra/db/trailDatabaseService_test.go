package tkInfraDb

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Note: Setup/teardown are intentionally inline — test independence
// requires each file to own its preconditions, even if it duplicates code.

func TestNewTrailDatabaseService(test *testing.T) {
	unixAbsoluteFilePathMaxLength := 4096
	overlongFilePath := "/tmp/" + strings.Repeat(
		"a", unixAbsoluteFilePathMaxLength+1,
	) + ".db"

	testCases := []struct {
		name              string
		envValue          string
		shouldUseTempPath bool
		expectedError     string
	}{
		{
			name:          "EnvVarEmpty",
			envValue:      "",
			expectedError: errTrailDatabaseFilePathNotSet,
		},
		{
			name:          "OverlongPath",
			envValue:      overlongFilePath,
			expectedError: errTrailDatabaseFilePathNotValid,
		},
		{
			name:              "ValidPath",
			shouldUseTempPath: true,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			dbFilePath := filepath.Join(subtest.TempDir(), "trail.db")
			envValue := testCase.envValue
			if testCase.shouldUseTempPath {
				envValue = dbFilePath
			}
			subtest.Setenv(TrailDatabaseFilePathEnvVarName, envValue)

			dbSvc, err := NewTrailDatabaseService([]any{})
			if testCase.expectedError != "" {
				if err == nil {
					subtest.Fatalf("MissingExpectedError: %s", testCase.expectedError)
				}
				if !strings.HasPrefix(err.Error(), testCase.expectedError) {
					subtest.Fatalf(
						"UnexpectedErrorMessage: '%s' vs '%s'",
						err.Error(), testCase.expectedError,
					)
				}
				return
			}

			if err != nil {
				subtest.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if dbSvc == nil {
				subtest.Fatal("ServiceIsNil")
			}
			if dbSvc.Handler == nil {
				subtest.Fatal("HandlerIsNil")
			}

			_, statErr := os.Stat(dbFilePath)
			if statErr != nil {
				subtest.Fatalf("DatabaseFileNotCreated: %s", dbFilePath)
			}
		})
	}
}

func TestNewTrailDatabaseServiceWithExtraModels(test *testing.T) {
	type testValidModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"column:name"`
	}

	testCases := []struct {
		name             string
		extraModels      []any
		expectedError    string
		expectedTablePtr any
	}{
		{
			name:        "NoExtraModels",
			extraModels: []any{},
		},
		{
			name:             "ValidExtraModel",
			extraModels:      []any{&testValidModel{}},
			expectedTablePtr: &testValidModel{},
		},
		{
			name:          "InvalidExtraModel",
			extraModels:   []any{"invalid_string"},
			expectedError: errTrailDatabaseMigrationError,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			dbFilePath := filepath.Join(subtest.TempDir(), "trail.db")
			subtest.Setenv(TrailDatabaseFilePathEnvVarName, dbFilePath)

			dbSvc, err := NewTrailDatabaseService(testCase.extraModels)
			if testCase.expectedError != "" {
				if err == nil {
					subtest.Fatalf("MissingExpectedError: %s", testCase.expectedError)
				}
				if !strings.HasPrefix(err.Error(), testCase.expectedError) {
					subtest.Fatalf(
						"UnexpectedErrorMessage: '%s' vs '%s'",
						err.Error(), testCase.expectedError,
					)
				}
				return
			}

			if err != nil {
				subtest.Fatalf("UnexpectedError: '%s'", err.Error())
			}

			if testCase.expectedTablePtr == nil {
				return
			}
			if !dbSvc.Handler.Migrator().HasTable(testCase.expectedTablePtr) {
				subtest.Fatalf("ExpectedTableMissing: %T", testCase.expectedTablePtr)
			}
		})
	}
}

func readIndexColumnNames(
	test *testing.T,
	dbSvc *TrailDatabaseService,
	indexName string,
) []string {
	test.Helper()

	columnNames := []string{}
	err := dbSvc.Handler.Raw(
		"SELECT name FROM pragma_index_info(?)", indexName,
	).Scan(&columnNames).Error
	if err != nil {
		test.Fatalf("ReadIndexColumnsFailed: %s", err.Error())
	}

	return columnNames
}

func TestTrailDatabaseServiceActivityRecordIndexes(test *testing.T) {
	dbFilePath := filepath.Join(test.TempDir(), "trail.db")
	test.Setenv(TrailDatabaseFilePathEnvVarName, dbFilePath)

	dbSvc, err := NewTrailDatabaseService([]any{})
	if err != nil {
		test.Fatalf("UnexpectedError: '%s'", err.Error())
	}

	testCases := []struct {
		name            string
		indexName       string
		expectedColumns []string
	}{
		{
			name:      "CreatedAt",
			indexName: "idx_activity_records_created_at",
			expectedColumns: []string{
				"created_at",
			},
		},
		{
			name:      "OperatorSriCreatedAt",
			indexName: "idx_activity_records_operator_sri_created_at",
			expectedColumns: []string{
				"operator_sri", "created_at",
			},
		},
		{
			name:      "CodeIpCreatedAt",
			indexName: "idx_activity_records_code_ip_created_at",
			expectedColumns: []string{
				"record_code", "operator_ip_address", "created_at",
			},
		},
		{
			name:      "AffectedResourceRecord",
			indexName: "idx_activity_records_affected_resources_record",
			expectedColumns: []string{
				"activity_record_id",
			},
		},
		{
			name:      "AffectedResourceResourceRecord",
			indexName: "idx_activity_records_affected_resources_resource_record",
			expectedColumns: []string{
				"system_resource_identifier", "activity_record_id",
			},
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			actualColumns := readIndexColumnNames(
				subtest, dbSvc, testCase.indexName,
			)
			if !slices.Equal(actualColumns, testCase.expectedColumns) {
				subtest.Errorf(
					"IndexColumnsMismatch: got='%v', want='%v'",
					actualColumns, testCase.expectedColumns,
				)
			}
		})
	}
}
