package tkPresentation

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

func TestUiRequestInputReader(test *testing.T) {
	newEchoContext := func(requestUri string) echo.Context {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, requestUri, nil)
		httpRecorder := httptest.NewRecorder()
		return echoInstance.NewContext(httpRequest, httpRecorder)
	}

	setOperatorContext := func(subtest *testing.T, echoContext echo.Context) {
		subtest.Helper()

		echoContext.Set("operatorAccountId", tkValueObject.AccountId(7))

		operatorIpAddress, err := tkValueObject.NewIpAddress("203.0.113.5")
		if err != nil {
			subtest.Fatalf("InvalidOperatorIpAddress: %v", err)
		}
		echoContext.Set("operatorIpAddress", operatorIpAddress)
	}

	assertValue := func(
		subtest *testing.T, requestInput map[string]any, key string,
		expectedValue any,
	) {
		subtest.Helper()

		actualValue, exists := requestInput[key]
		if !exists {
			subtest.Errorf("MissingKey: %s", key)
			return
		}
		if !reflect.DeepEqual(actualValue, expectedValue) {
			subtest.Errorf(
				"ValueMismatch: %s expected %v, got %v",
				key, expectedValue, actualValue,
			)
		}
	}

	assertAbsent := func(
		subtest *testing.T, requestInput map[string]any, key string,
	) {
		subtest.Helper()

		if _, exists := requestInput[key]; exists {
			subtest.Errorf("UnexpectedKey: %s", key)
		}
	}

	test.Run("ReadsDataTableAndCarouselFilters", func(subtest *testing.T) {
		echoContext := newEchoContext(
			"/?pageNumber=2&itemsPerPage=25&sortBy=name&sortDirection=desc" +
				"&lastSeenId=abc123&status=running&tag=alpha&tag=beta" +
				"&coresMin=10&coresMax=20&search=web",
		)
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{Key: "status"},
				{
					Key:            "tags",
					QueryParamName: "tag",
					Kind:           UiQueryParamFilterKindMulti,
				},
				{
					Key:            "cpu",
					QueryParamName: "cores",
					Kind:           UiQueryParamFilterKindRange,
				},
				{Key: "search"},
			},
			BaseInput: map[string]any{"withMetrics": true, "status": "base"},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(subtest, requestInput, "withMetrics", true)
		assertValue(subtest, reader.BaseInput, "withMetrics", true)
		assertValue(subtest, reader.BaseInput, "status", "base")
		assertValue(subtest, requestInput, "pageNumber", uint32(2))
		assertValue(subtest, requestInput, "itemsPerPage", uint16(25))
		assertValue(
			subtest, requestInput, "sortBy",
			tkValueObject.PaginationSortBy("name"),
		)
		assertValue(
			subtest, requestInput, "sortDirection",
			tkValueObject.PaginationSortDirection("desc"),
		)
		assertValue(
			subtest, requestInput, "lastSeenId",
			tkValueObject.PaginationLastSeenId("abc123"),
		)
		assertValue(subtest, requestInput, "status", "running")
		assertValue(
			subtest, requestInput, "tags", []string{"alpha", "beta"},
		)
		assertValue(
			subtest, requestInput, "cpu",
			map[string]string{"min": "10", "max": "20"},
		)
		assertValue(subtest, requestInput, "search", "web")
		assertValue(
			subtest, requestInput, "operatorAccountId",
			tkValueObject.AccountId(7),
		)
		assertValue(
			subtest, requestInput, "operatorIpAddress",
			tkValueObject.IpAddress("203.0.113.5"),
		)
		assertValue(subtest, requestInput, "operatorSri", nil)
	})

	test.Run("ReadsOperatorSriFromEchoContext", func(subtest *testing.T) {
		echoContext := newEchoContext("/?status=running")
		setOperatorContext(subtest, echoContext)

		operatorSri, err := tkValueObject.NewSystemResourceIdentifier(
			"sri://0:test/operator",
		)
		if err != nil {
			subtest.Fatalf("InvalidOperatorSri: %v", err)
		}
		echoContext.Set("operatorSri", operatorSri)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{{Key: "status"}},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(
			subtest, requestInput, "operatorSri", operatorSri,
		)
	})

	test.Run("FallsBackToKeyAsQueryParamName", func(subtest *testing.T) {
		echoContext := newEchoContext("/?status=running")
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{{Key: "status"}},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(subtest, requestInput, "status", "running")
	})

	test.Run("SkipsBlankFilterValues", func(subtest *testing.T) {
		echoContext := newEchoContext(
			"/?status=&tag=&coresMin=&coresMax=",
		)
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{Key: "status"},
				{
					Key:  "tags",
					Kind: UiQueryParamFilterKindMulti,
				},
				{
					Key:  "cpu",
					Kind: UiQueryParamFilterKindRange,
				},
			},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertAbsent(subtest, requestInput, "status")
		assertAbsent(subtest, requestInput, "tags")
		assertAbsent(subtest, requestInput, "cpu")
	})

	test.Run("ReadsOneSidedRange", func(subtest *testing.T) {
		echoContext := newEchoContext("/?coresMin=10")
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{
					Key:            "cpu",
					QueryParamName: "cores",
					Kind:           UiQueryParamFilterKindRange,
				},
			},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(
			subtest, requestInput, "cpu",
			map[string]string{"min": "10", "max": ""},
		)
	})

	test.Run("ReadsMaxOnlyRange", func(subtest *testing.T) {
		echoContext := newEchoContext("/?coresMax=20")
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{
					Key:            "cpu",
					QueryParamName: "cores",
					Kind:           UiQueryParamFilterKindRange,
				},
			},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(
			subtest, requestInput, "cpu",
			map[string]string{"min": "", "max": "20"},
		)
	})

	test.Run("IgnoresUnconfiguredAndPrefixedParams", func(subtest *testing.T) {
		echoContext := newEchoContext(
			"/?accountsStatus=spoofed&status=running",
		)
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{{Key: "status"}},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(subtest, requestInput, "status", "running")
		assertAbsent(subtest, requestInput, "accountsStatus")
	})

	test.Run("SkipsReservedFilterKeys", func(subtest *testing.T) {
		echoContext := newEchoContext(
			"/?pageNumber=2&sortBy=name&operatorAccountId=99&status=running",
		)
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{Key: "pageNumber"},
				{Key: "sortBy"},
				{Key: "operatorAccountId"},
				{Key: "status"},
			},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(subtest, requestInput, "pageNumber", uint32(2))
		assertValue(
			subtest, requestInput, "sortBy",
			tkValueObject.PaginationSortBy("name"),
		)
		assertValue(
			subtest, requestInput, "operatorAccountId",
			tkValueObject.AccountId(7),
		)
		assertValue(subtest, requestInput, "status", "running")
	})

	test.Run("SkipsReservedFilterKeysWhenPaginationDropsValue", func(subtest *testing.T) {
		echoContext := newEchoContext(
			"/?pageNumber=invalid&sortBy=bad@sort" +
				"&operatorAccountId=99&status=running",
		)
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{Key: "pageNumber"},
				{Key: "sortBy"},
				{Key: "operatorAccountId"},
				{Key: "status"},
			},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertAbsent(subtest, requestInput, "pageNumber")
		assertAbsent(subtest, requestInput, "sortBy")
		assertValue(
			subtest, requestInput, "operatorAccountId",
			tkValueObject.AccountId(7),
		)
		assertValue(subtest, requestInput, "status", "running")
	})

	test.Run("OverwritesSpoofedOperatorFields", func(subtest *testing.T) {
		echoContext := newEchoContext(
			"/?operatorAccountId=99&operatorSri=sri://0:test/spoofed" +
				"&operatorIpAddress=198.51.100.9&status=running",
		)
		setOperatorContext(subtest, echoContext)

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{
				{Key: "operatorAccountId"},
				{Key: "operatorSri"},
				{Key: "operatorIpAddress"},
				{Key: "status"},
			},
			BaseInput: map[string]any{
				"operatorAccountId": tkValueObject.AccountId(99),
				"operatorSri":       "sri://0:test/base-spoofed",
				"operatorIpAddress": tkValueObject.IpAddress("198.51.100.9"),
			},
		}

		requestInput, err := reader.Read(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		assertValue(
			subtest, requestInput, "operatorAccountId",
			tkValueObject.AccountId(7),
		)
		assertValue(
			subtest, requestInput, "operatorIpAddress",
			tkValueObject.IpAddress("203.0.113.5"),
		)
		assertValue(subtest, requestInput, "operatorSri", nil)
		assertValue(subtest, requestInput, "status", "running")
	})

	test.Run("MissingOperatorAccountIdReturnsError", func(subtest *testing.T) {
		echoContext := newEchoContext("/?status=running")

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{{Key: "status"}},
		}

		requestInput, err := reader.Read(echoContext)
		if err == nil || err.Error() != "OperatorAccountIdNotFound" {
			subtest.Fatalf(
				"ExpectedError: 'OperatorAccountIdNotFound', got: %v", err,
			)
		}
		if requestInput != nil {
			subtest.Errorf("ExpectedNilRequestInput")
		}
	})

	test.Run("WrongOperatorAccountIdTypeReturnsError", func(subtest *testing.T) {
		echoContext := newEchoContext("/?status=running")
		echoContext.Set("operatorAccountId", "7")

		reader := UiRequestInputReader{
			Filters: []UiQueryParamFilter{{Key: "status"}},
		}

		requestInput, err := reader.Read(echoContext)
		if err == nil || err.Error() != "OperatorAccountIdNotFound" {
			subtest.Fatalf(
				"ExpectedError: 'OperatorAccountIdNotFound', got: %v", err,
			)
		}
		if requestInput != nil {
			subtest.Errorf("ExpectedNilRequestInput")
		}
	})
}
