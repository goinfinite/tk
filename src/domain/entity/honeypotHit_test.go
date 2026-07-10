package entity

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewHoneypotHit(t *testing.T) {
	validIpAddress, _ := tkValueObject.NewIpAddress("192.168.1.1")
	validUrlPath, _ := tkValueObject.NewUrlPath("/wp-login.php")
	validPathClass := tkValueObject.HoneypotPathClassStaticVulnerability
	validFirstHitAt, _ := tkValueObject.NewUnixTime(1700000000)
	validCreatedAt, _ := tkValueObject.NewUnixTime(1700001000)

	testCaseStructs := []struct {
		testName          string
		requesterIp       tkValueObject.IpAddress
		honeypotPath      tkValueObject.UrlPath
		hitClass          tkValueObject.HoneypotPathClass
		hitCount          uint64
		firstHitAt        tkValueObject.UnixTime
		createdAt         tkValueObject.UnixTime
		expectedIp        string
		expectedPath      string
		expectedClass     string
		expectedCount     uint64
	}{
		{
			testName:      "ValidParams",
			requesterIp:   validIpAddress,
			honeypotPath:  validUrlPath,
			hitClass:      validPathClass,
			hitCount:      1,
			firstHitAt:    validFirstHitAt,
			createdAt:     validCreatedAt,
			expectedIp:    "192.168.1.1",
			expectedPath:  "/wp-login.php",
			expectedClass: "staticVulnerability",
			expectedCount: 1,
		},
		{
			testName:      "EmptyIpAddress",
			requesterIp:   tkValueObject.IpAddress(""),
			honeypotPath:  validUrlPath,
			hitClass:      validPathClass,
			hitCount:      1,
			firstHitAt:    validFirstHitAt,
			createdAt:     validCreatedAt,
			expectedIp:    "",
			expectedPath:  "/wp-login.php",
			expectedClass: "staticVulnerability",
			expectedCount: 1,
		},
		{
			testName:      "ZeroHitCount",
			requesterIp:   validIpAddress,
			honeypotPath:  validUrlPath,
			hitClass:      validPathClass,
			hitCount:      0,
			firstHitAt:    validFirstHitAt,
			createdAt:     validCreatedAt,
			expectedIp:    "192.168.1.1",
			expectedPath:  "/wp-login.php",
			expectedClass: "staticVulnerability",
			expectedCount: 0,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			hitEntity := NewHoneypotHit(
				testCase.requesterIp,
				testCase.honeypotPath,
				testCase.hitClass,
				testCase.hitCount,
				testCase.firstHitAt,
				testCase.createdAt,
			)

			actualIp := hitEntity.RequesterIpAddress.String()
			if actualIp != testCase.expectedIp {
				t.Errorf(
					"UnexpectedRequesterIpAddress: got '%s', expected '%s'",
					actualIp, testCase.expectedIp,
				)
			}

			actualPath := hitEntity.HoneypotPath.String()
			if actualPath != testCase.expectedPath {
				t.Errorf(
					"UnexpectedHoneypotPath: got '%s', expected '%s'",
					actualPath, testCase.expectedPath,
				)
			}

			actualClass := hitEntity.HitClass.String()
			if actualClass != testCase.expectedClass {
				t.Errorf(
					"UnexpectedHitClass: got '%s', expected '%s'",
					actualClass, testCase.expectedClass,
				)
			}

			if hitEntity.HitCount != testCase.expectedCount {
				t.Errorf(
					"UnexpectedHitCount: got %d, expected %d",
					hitEntity.HitCount, testCase.expectedCount,
				)
			}

			if hitEntity.FirstHitAt.Int64() != testCase.firstHitAt.Int64() {
				t.Errorf(
					"UnexpectedFirstHitAt: got %d, expected %d",
					hitEntity.FirstHitAt.Int64(),
					testCase.firstHitAt.Int64(),
				)
			}

			if hitEntity.CreatedAt.Int64() != testCase.createdAt.Int64() {
				t.Errorf(
					"UnexpectedCreatedAt: got %d, expected %d",
					hitEntity.CreatedAt.Int64(),
					testCase.createdAt.Int64(),
				)
			}
		})
	}
}
