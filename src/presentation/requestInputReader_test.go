package tkPresentation

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestStringDotNotationToHierarchicalMap(t *testing.T) {
	requestInputReader := RequestInputReader{}

	testCases := []struct {
		name           string
		initialMap     map[string]any
		keys           []string
		value          string
		expectedResult map[string]any
	}{
		{
			name:       "SingleKey",
			initialMap: map[string]any{},
			keys:       []string{"name"},
			value:      "test",
			expectedResult: map[string]any{
				"name": "test",
			},
		},
		{
			name:       "TwoLevelNesting",
			initialMap: map[string]any{},
			keys:       []string{"user", "name"},
			value:      "john",
			expectedResult: map[string]any{
				"user": map[string]any{
					"name": "john",
				},
			},
		},
		{
			name:       "ThreeLevelNesting",
			initialMap: map[string]any{},
			keys:       []string{"user", "address", "city"},
			value:      "NYC",
			expectedResult: map[string]any{
				"user": map[string]any{
					"address": map[string]any{
						"city": "NYC",
					},
				},
			},
		},
		{
			name: "AddToExistingMap",
			initialMap: map[string]any{
				"user": map[string]any{
					"name": "john",
				},
			},
			keys:  []string{"user", "age"},
			value: "30",
			expectedResult: map[string]any{
				"user": map[string]any{
					"name": "john",
					"age":  "30",
				},
			},
		},
		{
			name: "KeyConflictScalarToNested",
			initialMap: map[string]any{
				"user": "simpleString",
			},
			keys:  []string{"user", "name"},
			value: "john",
			expectedResult: map[string]any{
				"user": "simpleString",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualHierarchicalMap := requestInputReader.StringDotNotationToHierarchicalMap(
				testCase.initialMap, testCase.keys, testCase.value,
			)

			actualJson, err := json.Marshal(actualHierarchicalMap)
			if err != nil {
				t.Fatalf("MarshalActualMapFailed: %v", err)
			}

			expectedJson, err := json.Marshal(testCase.expectedResult)
			if err != nil {
				t.Fatalf("MarshalExpectedMapFailed: %v", err)
			}

			if string(actualJson) != string(expectedJson) {
				t.Errorf(
					"MapMismatch\nExpected: %s\nGot: %s",
					string(expectedJson), string(actualJson),
				)
			}
		})
	}
}

func TestFormUrlEncodedDataProcessor(t *testing.T) {
	requestInputReader := RequestInputReader{}

	testCases := []struct {
		name           string
		formData       map[string][]string
		expectedResult map[string]any
	}{
		{
			name: "SimpleKeyValue",
			formData: map[string][]string{
				"name":  {"john"},
				"email": {"john@example.com"},
			},
			expectedResult: map[string]any{
				"name":  "john",
				"email": "john@example.com",
			},
		},
		{
			name: "MultipleValues",
			formData: map[string][]string{
				"tags": {"tag1", "tag2", "tag3"},
			},
			expectedResult: map[string]any{
				"tags": []string{"tag1", "tag2", "tag3"},
			},
		},
		{
			name: "DotNotationKey",
			formData: map[string][]string{
				"user.name":  {"john"},
				"user.email": {"john@example.com"},
			},
			expectedResult: map[string]any{
				"user": map[string]any{
					"name":  "john",
					"email": "john@example.com",
				},
			},
		},
		{
			name: "MixedKeysAndDotNotation",
			formData: map[string][]string{
				"name":       {"john"},
				"user.email": {"john@example.com"},
			},
			expectedResult: map[string]any{
				"name": "john",
				"user": map[string]any{
					"email": "john@example.com",
				},
			},
		},
		{
			name: "EmptyValues",
			formData: map[string][]string{
				"empty": {},
				"name":  {"john"},
			},
			expectedResult: map[string]any{
				"name": "john",
			},
		},
		{
			name: "SingleDotAtEnd",
			formData: map[string][]string{
				"key.": {"value"},
			},
			expectedResult: map[string]any{
				"key": map[string]any{
					"": "value",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			emptyRequestBody := map[string]any{}
			processedRequestBody := requestInputReader.FormUrlEncodedDataProcessor(
				emptyRequestBody, testCase.formData,
			)

			actualJson, err := json.Marshal(processedRequestBody)
			if err != nil {
				t.Fatalf("MarshalActualMapFailed: %v", err)
			}

			expectedJson, err := json.Marshal(testCase.expectedResult)
			if err != nil {
				t.Fatalf("MarshalExpectedMapFailed: %v", err)
			}

			if string(actualJson) != string(expectedJson) {
				t.Errorf(
					"MapMismatch\nExpected: %s\nGot: %s",
					string(expectedJson), string(actualJson),
				)
			}
		})
	}
}

func TestMultipartFilesProcessor(t *testing.T) {
	requestInputReader := RequestInputReader{}

	t.Run("SingleFilePerKey", func(t *testing.T) {
		uploadedFilesByKey := map[string][]*multipart.FileHeader{
			"file1": {
				{Filename: "test1.txt"},
			},
			"file2": {
				{Filename: "test2.txt"},
			},
		}

		processedFiles := requestInputReader.MultipartFilesProcessor(uploadedFilesByKey)
		if len(processedFiles) != 2 {
			t.Errorf("ExpectedTwoFilesButGot: %d", len(processedFiles))
		}

		if processedFiles["file1"].Filename != "test1.txt" {
			t.Errorf("File1NameMismatch: %s", processedFiles["file1"].Filename)
		}

		if processedFiles["file2"].Filename != "test2.txt" {
			t.Errorf("File2NameMismatch: %s", processedFiles["file2"].Filename)
		}
	})

	t.Run("MultipleFilesPerKey", func(t *testing.T) {
		uploadedFilesByKey := map[string][]*multipart.FileHeader{
			"files": {
				{Filename: "test1.txt"},
				{Filename: "test2.txt"},
				{Filename: "test3.txt"},
			},
		}

		processedFiles := requestInputReader.MultipartFilesProcessor(uploadedFilesByKey)

		if len(processedFiles) != 3 {
			t.Errorf("ExpectedThreeFilesButGot: %d", len(processedFiles))
		}

		if processedFiles["files_0"].Filename != "test1.txt" {
			t.Errorf("Files0NameMismatch: %s", processedFiles["files_0"].Filename)
		}

		if processedFiles["files_1"].Filename != "test2.txt" {
			t.Errorf("Files1NameMismatch: %s", processedFiles["files_1"].Filename)
		}

		if processedFiles["files_2"].Filename != "test3.txt" {
			t.Errorf("Files2NameMismatch: %s", processedFiles["files_2"].Filename)
		}
	})

	t.Run("MixedSingleAndMultipleFiles", func(t *testing.T) {
		uploadedFilesByKey := map[string][]*multipart.FileHeader{
			"single": {
				{Filename: "single.txt"},
			},
			"multiple": {
				{Filename: "multi1.txt"},
				{Filename: "multi2.txt"},
			},
		}

		processedFiles := requestInputReader.MultipartFilesProcessor(uploadedFilesByKey)
		if len(processedFiles) != 3 {
			t.Errorf("ExpectedThreeFilesButGot: %d", len(processedFiles))
		}

		if processedFiles["single"].Filename != "single.txt" {
			t.Errorf("SingleFileNameMismatch: %s", processedFiles["single"].Filename)
		}

		if processedFiles["multiple_0"].Filename != "multi1.txt" {
			t.Errorf("Multiple0NameMismatch: %s", processedFiles["multiple_0"].Filename)
		}

		if processedFiles["multiple_1"].Filename != "multi2.txt" {
			t.Errorf("Multiple1NameMismatch: %s", processedFiles["multiple_1"].Filename)
		}
	})
}

func TestRequestInputReader(t *testing.T) {
	requestInputReader := RequestInputReader{}

	t.Run("JsonContentType", func(t *testing.T) {
		testCases := []struct {
			name           string
			jsonBody       string
			expectedResult map[string]any
			expectError    bool
		}{
			{
				name:     "ValidJsonBody",
				jsonBody: `{"name": "john", "age": 30}`,
				expectedResult: map[string]any{
					"name": "john",
					"age":  float64(30),
				},
				expectError: false,
			},
			{
				name:        "InvalidJsonBody",
				jsonBody:    `{invalid json}`,
				expectError: true,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				echoInstance := echo.New()
				httpRequest := httptest.NewRequest(
					http.MethodPost, "/", strings.NewReader(testCase.jsonBody),
				)
				httpRequest.Header.Set("Content-Type", "application/json")
				httpRecorder := httptest.NewRecorder()
				echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

				parsedRequestBody, err := requestInputReader.Reader(echoContext)
				if testCase.expectError && err == nil {
					t.Errorf("MissingExpectedError")
				}

				if !testCase.expectError && err != nil {
					t.Errorf("UnexpectedError: %v", err)
				}

				if !testCase.expectError {
					for key, expectedValue := range testCase.expectedResult {
						if parsedRequestBody[key] != expectedValue {
							t.Errorf(
								"ValueMismatchForKey %s: expected %v, got %v",
								key, expectedValue, parsedRequestBody[key],
							)
						}
					}

					if parsedRequestBody["operatorIpAddress"] == nil {
						t.Errorf("OperatorIpAddressNotSet")
					}
				}
			})
		}
	})

	t.Run("FormUrlEncodedContentType", func(t *testing.T) {
		echoInstance := echo.New()
		formData := url.Values{}
		formData.Set("name", "john")
		formData.Set("email", "john@example.com")
		formData.Set("user.age", "30")

		httpRequest := httptest.NewRequest(
			http.MethodPost, "/", strings.NewReader(formData.Encode()),
		)
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["name"] != "john" {
			t.Errorf("NameMismatch: expected john, got %v", parsedRequestBody["name"])
		}

		if parsedRequestBody["email"] != "john@example.com" {
			t.Errorf(
				"EmailMismatch: expected john@example.com, got %v",
				parsedRequestBody["email"],
			)
		}

		userMap, isMap := parsedRequestBody["user"].(map[string]any)
		if !isMap {
			t.Errorf("UserNotAMap: %v", parsedRequestBody["user"])
		}

		if userMap["age"] != "30" {
			t.Errorf("AgeMismatch: expected 30, got %v", userMap["age"])
		}
	})

	t.Run("MultipartFormDataContentType", func(t *testing.T) {
		echoInstance := echo.New()

		multipartBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(multipartBody)
		err := multipartWriter.WriteField("name", "john")
		if err != nil {
			t.Fatalf("WriteFieldNameFailed: %v", err)
		}

		err = multipartWriter.WriteField("email", "john@example.com")
		if err != nil {
			t.Fatalf("WriteFieldEmailFailed: %v", err)
		}

		fileWriter, err := multipartWriter.CreateFormFile("avatar", "avatar.jpg")
		if err != nil {
			t.Fatalf("CreateFormFileFailed: %v", err)
		}

		_, err = io.WriteString(fileWriter, "fake image content")
		if err != nil {
			t.Fatalf("WriteFileContentFailed: %v", err)
		}

		multipartWriter.Close()

		httpRequest := httptest.NewRequest(http.MethodPost, "/", multipartBody)
		httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["name"] != "john" {
			t.Errorf("NameMismatch: expected john, got %v", parsedRequestBody["name"])
		}

		if parsedRequestBody["email"] != "john@example.com" {
			t.Errorf(
				"EmailMismatch: expected john@example.com, got %v",
				parsedRequestBody["email"],
			)
		}

		uploadedFiles, hasFiles := parsedRequestBody["files"].(map[string]*multipart.FileHeader)
		if !hasFiles {
			t.Errorf("FilesNotFound")
		}

		if uploadedFiles["avatar"].Filename != "avatar.jpg" {
			t.Errorf(
				"AvatarFilenameMismatch: expected avatar.jpg, got %s",
				uploadedFiles["avatar"].Filename,
			)
		}
	})

	t.Run("MultipartFormDataWithMultipleFiles", func(t *testing.T) {
		echoInstance := echo.New()

		multipartBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(multipartBody)

		for fileIndex := range 3 {
			fileName := "doc" + strconv.Itoa(fileIndex) + ".pdf"
			fileWriter, err := multipartWriter.CreateFormFile("documents", fileName)
			if err != nil {
				t.Fatalf("CreateFormFileFailed: %v", err)
			}

			_, err = io.WriteString(fileWriter, "fake pdf content")
			if err != nil {
				t.Fatalf("WriteFileContentFailed: %v", err)
			}
		}

		multipartWriter.Close()

		httpRequest := httptest.NewRequest(http.MethodPost, "/", multipartBody)
		httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		uploadedFiles, hasFiles := parsedRequestBody["files"].(map[string]*multipart.FileHeader)
		if !hasFiles {
			t.Errorf("FilesNotFound")
		}

		if len(uploadedFiles) != 3 {
			t.Errorf("ExpectedThreeFilesButGot: %d", len(uploadedFiles))
		}

		if uploadedFiles["documents_0"].Filename != "doc0.pdf" {
			t.Errorf(
				"Documents0FilenameMismatch: expected doc0.pdf, got %s",
				uploadedFiles["documents_0"].Filename,
			)
		}
	})

	t.Run("InvalidContentType", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodPost, "/", nil)
		httpRequest.Header.Set("Content-Type", "text/plain")
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		_, err := requestInputReader.Reader(echoContext)
		if err == nil {
			t.Errorf("MissingExpectedError: InvalidContentType")
		}

		if !strings.Contains(err.Error(), "InvalidContentType") {
			t.Errorf("UnexpectedErrorMessage: %v", err)
		}
	})

	t.Run("QueryParamsAndPathParams", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/?search=test&limit=10", nil,
		)
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)
		echoContext.SetParamNames("id", "action")
		echoContext.SetParamValues("123", "update")

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["search"] != "test" {
			t.Errorf(
				"SearchParamMismatch: expected test, got %v",
				parsedRequestBody["search"],
			)
		}

		if parsedRequestBody["limit"] != "10" {
			t.Errorf(
				"LimitParamMismatch: expected 10, got %v",
				parsedRequestBody["limit"],
			)
		}

		if parsedRequestBody["id"] != "123" {
			t.Errorf(
				"IdParamMismatch: expected 123, got %v",
				parsedRequestBody["id"],
			)
		}

		if parsedRequestBody["action"] != "update" {
			t.Errorf(
				"ActionParamMismatch: expected update, got %v",
				parsedRequestBody["action"],
			)
		}
	})

	t.Run("OperatorAccountIdFromContext", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)
		echoContext.Set("operatorAccountId", uint64(42))

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["operatorAccountId"] != uint64(42) {
			t.Errorf(
				"OperatorAccountIdMismatch: expected 42, got %v",
				parsedRequestBody["operatorAccountId"],
			)
		}

		if parsedRequestBody["operatorIpAddress"] == nil {
			t.Errorf("OperatorIpAddressNotSet")
		}
	})

	t.Run("OperatorIpAddressAlwaysSet", func(t *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRequest.RemoteAddr = "192.168.1.100:12345"
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["operatorIpAddress"] == nil {
			t.Errorf("OperatorIpAddressNotSet")
		}
	})
}
