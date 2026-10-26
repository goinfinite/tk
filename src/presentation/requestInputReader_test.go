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

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

func TestStringDotNotationToHierarchicalMap(test *testing.T) {
	requestInputReader := ApiRequestInputReader{}

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
		test.Run(testCase.name, func(subtest *testing.T) {
			actualHierarchicalMap := requestInputReader.StringDotNotationToHierarchicalMap(
				testCase.initialMap, testCase.keys, testCase.value,
			)

			actualJson, err := json.Marshal(actualHierarchicalMap)
			if err != nil {
				subtest.Fatalf("MarshalActualMapFailed: %v", err)
			}

			expectedJson, err := json.Marshal(testCase.expectedResult)
			if err != nil {
				subtest.Fatalf("MarshalExpectedMapFailed: %v", err)
			}

			if string(actualJson) != string(expectedJson) {
				subtest.Errorf(
					"MapMismatch\nExpected: %s\nGot: %s",
					string(expectedJson), string(actualJson),
				)
			}
		})
	}
}

func TestFormUrlEncodedDataProcessor(test *testing.T) {
	requestInputReader := ApiRequestInputReader{}

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
		test.Run(testCase.name, func(subtest *testing.T) {
			emptyRequestBody := map[string]any{}
			processedRequestBody := requestInputReader.FormUrlEncodedDataProcessor(
				emptyRequestBody, testCase.formData,
			)

			actualJson, err := json.Marshal(processedRequestBody)
			if err != nil {
				subtest.Fatalf("MarshalActualMapFailed: %v", err)
			}

			expectedJson, err := json.Marshal(testCase.expectedResult)
			if err != nil {
				subtest.Fatalf("MarshalExpectedMapFailed: %v", err)
			}

			if string(actualJson) != string(expectedJson) {
				subtest.Errorf(
					"MapMismatch\nExpected: %s\nGot: %s",
					string(expectedJson), string(actualJson),
				)
			}
		})
	}
}

func TestApiRequestInputReader(test *testing.T) {
	requestInputReader := ApiRequestInputReader{}

	test.Run("JsonContentType", func(subtest *testing.T) {
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
			subtest.Run(testCase.name, func(subtest *testing.T) {
				echoInstance := echo.New()
				httpRequest := httptest.NewRequest(
					http.MethodPost, "/", strings.NewReader(testCase.jsonBody),
				)
				httpRequest.Header.Set("Content-Type", "application/json")
				httpRecorder := httptest.NewRecorder()
				echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

				parsedRequestBody, err := requestInputReader.Reader(echoContext)
				if testCase.expectError && err == nil {
					subtest.Errorf("MissingExpectedError")
				}

				if !testCase.expectError && err != nil {
					subtest.Errorf("UnexpectedError: %v", err)
				}

				if !testCase.expectError {
					for key, expectedValue := range testCase.expectedResult {
						if parsedRequestBody[key] != expectedValue {
							subtest.Errorf(
								"ValueMismatchForKey %s: expected %v, got %v",
								key, expectedValue, parsedRequestBody[key],
							)
						}
					}

					if parsedRequestBody["operatorIpAddress"] == nil {
						subtest.Errorf("OperatorIpAddressNotSet")
					}
				}
			})
		}
	})

	test.Run("FormUrlEncodedContentType", func(subtest *testing.T) {
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
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["name"] != "john" {
			subtest.Errorf("NameMismatch: expected john, got %v", parsedRequestBody["name"])
		}

		if parsedRequestBody["email"] != "john@example.com" {
			subtest.Errorf(
				"EmailMismatch: expected john@example.com, got %v",
				parsedRequestBody["email"],
			)
		}

		userMap, isMap := parsedRequestBody["user"].(map[string]any)
		if !isMap {
			subtest.Errorf("UserNotAMap: %v", parsedRequestBody["user"])
		}

		if userMap["age"] != "30" {
			subtest.Errorf("AgeMismatch: expected 30, got %v", userMap["age"])
		}
	})

	test.Run("MultipartFormDataContentType", func(subtest *testing.T) {
		echoInstance := echo.New()

		multipartBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(multipartBody)
		err := multipartWriter.WriteField("name", "john")
		if err != nil {
			subtest.Fatalf("WriteFieldNameFailed: %v", err)
		}

		err = multipartWriter.WriteField("email", "john@example.com")
		if err != nil {
			subtest.Fatalf("WriteFieldEmailFailed: %v", err)
		}

		fileWriter, err := multipartWriter.CreateFormFile("avatar", "avatar.jpg")
		if err != nil {
			subtest.Fatalf("CreateFormFileFailed: %v", err)
		}

		_, err = io.WriteString(fileWriter, "fake image content")
		if err != nil {
			subtest.Fatalf("WriteFileContentFailed: %v", err)
		}

		_ = multipartWriter.Close()

		httpRequest := httptest.NewRequest(http.MethodPost, "/", multipartBody)
		httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["name"] != "john" {
			subtest.Errorf("NameMismatch: expected john, got %v", parsedRequestBody["name"])
		}

		if parsedRequestBody["email"] != "john@example.com" {
			subtest.Errorf(
				"EmailMismatch: expected john@example.com, got %v",
				parsedRequestBody["email"],
			)
		}

		uploadedFiles, hasFiles := parsedRequestBody["files"].(map[string][]*multipart.FileHeader)
		if !hasFiles {
			subtest.Errorf("FilesNotFound")
		}

		avatarFiles := uploadedFiles["avatar"]
		if len(avatarFiles) != 1 {
			subtest.Fatalf("ExpectedOneAvatarFileButGot: %d", len(avatarFiles))
		}

		if avatarFiles[0].Filename != "avatar.jpg" {
			subtest.Errorf(
				"AvatarFilenameMismatch: expected avatar.jpg, got %s",
				avatarFiles[0].Filename,
			)
		}
	})

	test.Run("MultipartFormDataWithMultipleFiles", func(subtest *testing.T) {
		echoInstance := echo.New()

		multipartBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(multipartBody)

		for fileIndex := range 3 {
			fileName := "doc" + strconv.Itoa(fileIndex) + ".pdf"
			fileWriter, err := multipartWriter.CreateFormFile("documents", fileName)
			if err != nil {
				subtest.Fatalf("CreateFormFileFailed: %v", err)
			}

			_, err = io.WriteString(fileWriter, "fake pdf content")
			if err != nil {
				subtest.Fatalf("WriteFileContentFailed: %v", err)
			}
		}

		_ = multipartWriter.Close()

		httpRequest := httptest.NewRequest(http.MethodPost, "/", multipartBody)
		httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		uploadedFiles, hasFiles := parsedRequestBody["files"].(map[string][]*multipart.FileHeader)
		if !hasFiles {
			subtest.Errorf("FilesNotFound")
		}

		documentFiles := uploadedFiles["documents"]
		if len(documentFiles) != 3 {
			subtest.Fatalf("ExpectedThreeDocumentsButGot: %d", len(documentFiles))
		}

		for fileIndex, documentFile := range documentFiles {
			expectedFileName := "doc" + strconv.Itoa(fileIndex) + ".pdf"
			if documentFile.Filename != expectedFileName {
				subtest.Errorf(
					"DocumentFilenameMismatchAtIndex %d: expected %s, got %s",
					fileIndex, expectedFileName, documentFile.Filename,
				)
			}
		}
	})

	test.Run("MultipartFormDataWithLiteralIndexedFieldName", func(subtest *testing.T) {
		echoInstance := echo.New()

		multipartBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(multipartBody)

		fileWriter, err := multipartWriter.CreateFormFile("archiveFiles_0", "literal.txt")
		if err != nil {
			subtest.Fatalf("CreateFormFileFailed: %v", err)
		}

		_, err = io.WriteString(fileWriter, "fake content")
		if err != nil {
			subtest.Fatalf("WriteFileContentFailed: %v", err)
		}

		_ = multipartWriter.Close()

		httpRequest := httptest.NewRequest(http.MethodPost, "/", multipartBody)
		httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		uploadedFiles, hasFiles := parsedRequestBody["files"].(map[string][]*multipart.FileHeader)
		if !hasFiles {
			subtest.Fatalf("FilesNotFound")
		}

		literalFiles := uploadedFiles["archiveFiles_0"]
		if len(literalFiles) != 1 {
			subtest.Fatalf("ExpectedOneLiteralFileButGot: %d", len(literalFiles))
		}

		if literalFiles[0].Filename != "literal.txt" {
			subtest.Errorf(
				"LiteralFilenameMismatch: expected literal.txt, got %s",
				literalFiles[0].Filename,
			)
		}
	})

	test.Run("MultipartFormDataWithoutFiles", func(subtest *testing.T) {
		echoInstance := echo.New()

		multipartBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(multipartBody)
		err := multipartWriter.WriteField("name", "john")
		if err != nil {
			subtest.Fatalf("WriteFieldNameFailed: %v", err)
		}

		_ = multipartWriter.Close()

		httpRequest := httptest.NewRequest(http.MethodPost, "/", multipartBody)
		httpRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if _, hasFiles := parsedRequestBody["files"]; hasFiles {
			subtest.Errorf("UnexpectedFilesKey")
		}
	})

	test.Run("InvalidContentType", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodPost, "/", nil)
		httpRequest.Header.Set("Content-Type", "text/plain")
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		_, err := requestInputReader.Reader(echoContext)
		if err == nil {
			subtest.Errorf("MissingExpectedError: InvalidContentType")
		}

		if !strings.Contains(err.Error(), "InvalidContentType") {
			subtest.Errorf("UnexpectedErrorMessage: %v", err)
		}
	})

	test.Run("QueryParamsAndPathParams", func(subtest *testing.T) {
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
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["search"] != "test" {
			subtest.Errorf(
				"SearchParamMismatch: expected test, got %v",
				parsedRequestBody["search"],
			)
		}

		if parsedRequestBody["limit"] != "10" {
			subtest.Errorf(
				"LimitParamMismatch: expected 10, got %v",
				parsedRequestBody["limit"],
			)
		}

		if parsedRequestBody["id"] != "123" {
			subtest.Errorf(
				"IdParamMismatch: expected 123, got %v",
				parsedRequestBody["id"],
			)
		}

		if parsedRequestBody["action"] != "update" {
			subtest.Errorf(
				"ActionParamMismatch: expected update, got %v",
				parsedRequestBody["action"],
			)
		}
	})

	test.Run("OperatorAccountIdFromContext", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)
		echoContext.Set("operatorAccountId", tkValueObject.AccountId(42))

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["operatorAccountId"] != tkValueObject.AccountId(42) {
			subtest.Errorf(
				"OperatorAccountIdMismatch: expected 42, got %v",
				parsedRequestBody["operatorAccountId"],
			)
		}

		if parsedRequestBody["operatorIpAddress"] == nil {
			subtest.Errorf("OperatorIpAddressNotSet")
		}
	})

	test.Run("OperatorIpAddressAlwaysSet", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRequest.RemoteAddr = "192.168.1.100:12345"
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		parsedRequestBody, err := requestInputReader.Reader(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if parsedRequestBody["operatorIpAddress"] == nil {
			subtest.Errorf("OperatorIpAddressNotSet")
		}
	})
}
