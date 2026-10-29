package tkUseCase

import (
	"log/slog"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestActivityRecordLogAttributes(test *testing.T) {
	logAttributesMap := func(
		subtest *testing.T, attributes []any,
	) map[string]string {
		subtest.Helper()

		attributesMap := make(map[string]string, len(attributes))
		for _, rawAttribute := range attributes {
			attribute, isAttribute := rawAttribute.(slog.Attr)
			if !isAttribute {
				subtest.Fatalf("UnexpectedAttributeType: %T", rawAttribute)
			}
			attributesMap[attribute.Key] = attribute.Value.String()
		}

		return attributesMap
	}

	recordId, err := tkValueObject.NewActivityRecordId(123)
	if err != nil {
		test.Fatalf("RecordIdCreationFailed: %v", err)
	}
	recordLevel, err := tkValueObject.NewActivityRecordLevel("SECURITY")
	if err != nil {
		test.Fatalf("RecordLevelCreationFailed: %v", err)
	}
	recordCode, err := tkValueObject.NewActivityRecordCode("TEST_CODE")
	if err != nil {
		test.Fatalf("RecordCodeCreationFailed: %v", err)
	}
	resourceOne, err := tkValueObject.NewSystemResourceIdentifier(
		"sri://0:test/resource1",
	)
	if err != nil {
		test.Fatalf("ResourceOneCreationFailed: %v", err)
	}
	resourceTwo, err := tkValueObject.NewSystemResourceIdentifier(
		"sri://0:test/resource2",
	)
	if err != nil {
		test.Fatalf("ResourceTwoCreationFailed: %v", err)
	}
	operatorSri, err := tkValueObject.NewSystemResourceIdentifier(
		"sri://0:test/operator",
	)
	if err != nil {
		test.Fatalf("OperatorSriCreationFailed: %v", err)
	}
	operatorIpAddress, err := tkValueObject.NewIpAddress("192.168.1.1")
	if err != nil {
		test.Fatalf("OperatorIpAddressCreationFailed: %v", err)
	}
	createdBeforeAt, err := tkValueObject.NewUnixTime(1234567890)
	if err != nil {
		test.Fatalf("CreatedBeforeAtCreationFailed: %v", err)
	}
	createdAfterAt, err := tkValueObject.NewUnixTime(1234567891)
	if err != nil {
		test.Fatalf("CreatedAfterAtCreationFailed: %v", err)
	}
	sortBy, err := tkValueObject.NewPaginationSortBy("createdAt")
	if err != nil {
		test.Fatalf("SortByCreationFailed: %v", err)
	}
	sortDirection, err := tkValueObject.NewPaginationSortDirection("desc")
	if err != nil {
		test.Fatalf("SortDirectionCreationFailed: %v", err)
	}
	lastSeenId, err := tkValueObject.NewPaginationLastSeenId("abc123")
	if err != nil {
		test.Fatalf("LastSeenIdCreationFailed: %v", err)
	}
	pagination := tkDto.Pagination{
		PageNumber:    2,
		ItemsPerPage:  25,
		SortBy:        &sortBy,
		SortDirection: &sortDirection,
		LastSeenId:    &lastSeenId,
	}

	test.Run("AllFields", func(subtest *testing.T) {
		attributes := activityRecordLogAttributesBuilder(
			activityRecordLogAttributesInput{
				Pagination:        &pagination,
				RecordId:          &recordId,
				RecordLevel:       &recordLevel,
				RecordCode:        &recordCode,
				AffectedResources: []tkValueObject.SystemResourceIdentifier{resourceOne, resourceTwo},
				OperatorSri:       &operatorSri,
				OperatorIpAddress: &operatorIpAddress,
				CreatedBeforeAt:   &createdBeforeAt,
				CreatedAfterAt:    &createdAfterAt,
			},
		)
		attributesMap := logAttributesMap(subtest, attributes)

		expectedAttributes := map[string]string{
			"affectedResourcesCount": "2",
			"affectedResources":      "sri://0:test/resource1,sri://0:test/resource2",
			"pageNumber":             "2",
			"itemsPerPage":           "25",
			"sortBy":                 "createdAt",
			"sortDirection":          "desc",
			"lastSeenId":             "abc123",
			"recordId":               "123",
			"recordLevel":            "SECURITY",
			"recordCode":             "TEST_CODE",
			"operatorSri":            "sri://0:test/operator",
			"operatorIpAddress":      "192.168.1.1",
			"createdBeforeAt":        "1234567890",
			"createdAfterAt":         "1234567891",
		}
		for key, expectedValue := range expectedAttributes {
			actualValue, exists := attributesMap[key]
			if !exists {
				subtest.Errorf("MissingAttribute: %s", key)
				continue
			}
			if actualValue != expectedValue {
				subtest.Errorf(
					"AttributeValueMismatch: %s expected '%s', got '%s'",
					key, expectedValue, actualValue,
				)
			}
		}
	})

	test.Run("NilOptionalFields", func(subtest *testing.T) {
		attributes := activityRecordLogAttributesBuilder(
			activityRecordLogAttributesInput{},
		)
		attributesMap := logAttributesMap(subtest, attributes)

		if attributesMap["affectedResourcesCount"] != "0" {
			subtest.Errorf(
				"AffectedResourcesCountMismatch: expected '0', got '%s'",
				attributesMap["affectedResourcesCount"],
			)
		}
		optionalKeys := []string{
			"affectedResources",
			"pageNumber",
			"itemsPerPage",
			"sortBy",
			"sortDirection",
			"lastSeenId",
			"recordId",
			"recordLevel",
			"recordCode",
			"operatorSri",
			"operatorIpAddress",
			"createdBeforeAt",
			"createdAfterAt",
		}
		for _, key := range optionalKeys {
			if _, exists := attributesMap[key]; exists {
				subtest.Errorf("UnexpectedAttribute: %s", key)
			}
		}
	})
}
