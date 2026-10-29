package tkPresentation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

func TestPaginationQueryParamsInputReader(test *testing.T) {
	test.Run("ReadsPaginationKeysFromGetRequestWithoutContentType", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet,
			"/?pageNumber=2&itemsPerPage=25&sortBy=name&sortDirection=desc&lastSeenId=abc123",
			nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsInputReader(echoContext, nil)

		if requestParams["pageNumber"] != uint32(2) {
			subtest.Errorf("PageNumberMismatch: expected 2, got %v", requestParams["pageNumber"])
		}

		if requestParams["itemsPerPage"] != uint16(25) {
			subtest.Errorf(
				"ItemsPerPageMismatch: expected 25, got %v",
				requestParams["itemsPerPage"],
			)
		}

		if requestParams["sortBy"] != tkValueObject.PaginationSortBy("name") {
			subtest.Errorf("SortByMismatch: expected name, got %v", requestParams["sortBy"])
		}

		if requestParams["sortDirection"] != tkValueObject.PaginationSortDirection("desc") {
			subtest.Errorf(
				"SortDirectionMismatch: expected desc, got %v",
				requestParams["sortDirection"],
			)
		}

		if requestParams["lastSeenId"] != tkValueObject.PaginationLastSeenId("abc123") {
			subtest.Errorf(
				"LastSeenIdMismatch: expected abc123, got %v",
				requestParams["lastSeenId"],
			)
		}
	})

	test.Run("DropsMalformedPaginationValues", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet,
			"/?pageNumber=invalid&itemsPerPage=0&sortBy=bad@sort&sortDirection=ascending&lastSeenId=with.dot",
			nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsInputReader(echoContext, nil)

		for _, paginationKey := range []string{
			"pageNumber", "itemsPerPage", "sortBy", "sortDirection", "lastSeenId",
		} {
			if _, hasKey := requestParams[paginationKey]; hasKey {
				subtest.Errorf("UnexpectedKey: %s", paginationKey)
			}
		}
	})

	test.Run("MalformedSortByFallsBackToPaginationDtoParserDefault", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?sortBy=bad@sort", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsInputReader(echoContext, nil)

		defaultPagination := tkDto.Pagination{PageNumber: 1, ItemsPerPage: 10}
		parsedPagination, err := PaginationDtoParser(defaultPagination, requestParams)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if parsedPagination.SortBy != nil {
			subtest.Errorf("UnexpectedSortBy: %v", parsedPagination.SortBy)
		}
	})

	test.Run("ReadsOptionalParamsWhenPresentAndNonEmpty", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?nature=app&name=&status=running", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		optionalParamNames := []string{"nature", "name", "status", "type"}
		requestParams := PaginationQueryParamsInputReader(echoContext, optionalParamNames)

		if requestParams["nature"] != "app" {
			subtest.Errorf("NatureMismatch: expected app, got %v", requestParams["nature"])
		}

		if requestParams["status"] != "running" {
			subtest.Errorf("StatusMismatch: expected running, got %v", requestParams["status"])
		}

		if _, hasName := requestParams["name"]; hasName {
			subtest.Errorf("UnexpectedEmptyNameParam")
		}

		if _, hasType := requestParams["type"]; hasType {
			subtest.Errorf("UnexpectedAbsentTypeParam")
		}
	})

	test.Run("IgnoresPaginationKeysListedAsOptional", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?pageNumber=2&sortBy=bad@sort&nature=app", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		optionalParamNames := []string{"pageNumber", "sortBy", "nature"}
		requestParams := PaginationQueryParamsInputReader(echoContext, optionalParamNames)

		if requestParams["pageNumber"] != uint32(2) {
			subtest.Errorf(
				"PageNumberMismatch: expected converted 2, got %v",
				requestParams["pageNumber"],
			)
		}

		if _, hasSortBy := requestParams["sortBy"]; hasSortBy {
			subtest.Errorf("UnexpectedMalformedSortBy: %v", requestParams["sortBy"])
		}

		if requestParams["nature"] != "app" {
			subtest.Errorf("NatureMismatch: expected app, got %v", requestParams["nature"])
		}
	})

	test.Run("ReadsOnlyPaginationKeysWithoutOptionalNames", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?pageNumber=3&nature=app", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsInputReader(echoContext, nil)

		if len(requestParams) != 1 {
			subtest.Fatalf("ExpectedOneParamButGot: %d", len(requestParams))
		}

		if requestParams["pageNumber"] != uint32(3) {
			subtest.Errorf("PageNumberMismatch: expected 3, got %v", requestParams["pageNumber"])
		}
	})
}
