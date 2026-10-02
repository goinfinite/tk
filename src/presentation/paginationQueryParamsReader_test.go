package tkPresentation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

func TestPaginationQueryParamsReader(t *testing.T) {
	t.Run("ReadsPaginationKeysFromGetRequestWithoutContentType", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet,
			"/?pageNumber=2&itemsPerPage=25&sortBy=name&sortDirection=desc&lastSeenId=abc123",
			nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsReader(echoContext, nil)

		if requestParams["pageNumber"] != uint32(2) {
			t.Errorf("PageNumberMismatch: expected 2, got %v", requestParams["pageNumber"])
		}

		if requestParams["itemsPerPage"] != uint16(25) {
			t.Errorf(
				"ItemsPerPageMismatch: expected 25, got %v",
				requestParams["itemsPerPage"],
			)
		}

		if requestParams["sortBy"] != tkValueObject.PaginationSortBy("name") {
			t.Errorf("SortByMismatch: expected name, got %v", requestParams["sortBy"])
		}

		if requestParams["sortDirection"] != tkValueObject.PaginationSortDirection("desc") {
			t.Errorf(
				"SortDirectionMismatch: expected desc, got %v",
				requestParams["sortDirection"],
			)
		}

		if requestParams["lastSeenId"] != tkValueObject.PaginationLastSeenId("abc123") {
			t.Errorf(
				"LastSeenIdMismatch: expected abc123, got %v",
				requestParams["lastSeenId"],
			)
		}
	})

	t.Run("DropsMalformedPaginationValues", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet,
			"/?pageNumber=invalid&itemsPerPage=0&sortBy=bad@sort&sortDirection=ascending&lastSeenId=with.dot",
			nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsReader(echoContext, nil)

		for _, paginationKey := range []string{
			"pageNumber", "itemsPerPage", "sortBy", "sortDirection", "lastSeenId",
		} {
			if _, hasKey := requestParams[paginationKey]; hasKey {
				t.Errorf("UnexpectedKey: %s", paginationKey)
			}
		}
	})

	t.Run("MalformedSortByFallsBackToPaginationParserDefault", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?sortBy=bad@sort", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsReader(echoContext, nil)

		defaultPagination := tkDto.Pagination{PageNumber: 1, ItemsPerPage: 10}
		parsedPagination, err := PaginationParser(defaultPagination, requestParams)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if parsedPagination.SortBy != nil {
			t.Errorf("UnexpectedSortBy: %v", parsedPagination.SortBy)
		}
	})

	t.Run("ReadsOptionalParamsWhenPresentAndNonEmpty", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?nature=app&name=&status=running", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		optionalParamNames := []string{"nature", "name", "status", "type"}
		requestParams := PaginationQueryParamsReader(echoContext, optionalParamNames)

		if requestParams["nature"] != "app" {
			t.Errorf("NatureMismatch: expected app, got %v", requestParams["nature"])
		}

		if requestParams["status"] != "running" {
			t.Errorf("StatusMismatch: expected running, got %v", requestParams["status"])
		}

		if _, hasName := requestParams["name"]; hasName {
			t.Errorf("UnexpectedEmptyNameParam")
		}

		if _, hasType := requestParams["type"]; hasType {
			t.Errorf("UnexpectedAbsentTypeParam")
		}
	})

	t.Run("ReadsOnlyPaginationKeysWithoutOptionalNames", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?pageNumber=3&nature=app", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		requestParams := PaginationQueryParamsReader(echoContext, nil)

		if len(requestParams) != 1 {
			t.Fatalf("ExpectedOneParamButGot: %d", len(requestParams))
		}

		if requestParams["pageNumber"] != uint32(3) {
			t.Errorf("PageNumberMismatch: expected 3, got %v", requestParams["pageNumber"])
		}
	})
}
