package tkPresentation

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type RequestInputReader struct {
}

func (reader RequestInputReader) StringDotNotationToHierarchicalMap(
	hierarchicalMap map[string]any, remainingKeys []string, finalValue string,
) map[string]any {
	if len(remainingKeys) == 1 {
		hierarchicalMap[remainingKeys[0]] = finalValue
		return hierarchicalMap
	}

	parentKey := remainingKeys[0]
	nextKeys := remainingKeys[1:]

	if _, exists := hierarchicalMap[parentKey]; !exists {
		hierarchicalMap[parentKey] = make(map[string]any)
	}

	hierarchicalMap[parentKey] = reader.StringDotNotationToHierarchicalMap(
		hierarchicalMap[parentKey].(map[string]any), nextKeys, finalValue,
	)

	return hierarchicalMap
}

func (reader RequestInputReader) FormUrlEncodedDataProcessor(
	requestBody map[string]any, formData map[string][]string,
) map[string]any {
	for formKey, formValues := range formData {
		if len(formValues) == 0 {
			continue
		}

		if len(formValues) > 1 {
			requestBody[formKey] = formValues
			continue
		}

		formValue := formValues[0]
		if !strings.Contains(formKey, ".") {
			requestBody[formKey] = formValue
			continue
		}

		keyParts := strings.Split(formKey, ".")
		if len(keyParts) < 2 {
			continue
		}

		requestBody = reader.StringDotNotationToHierarchicalMap(
			requestBody, keyParts, formValue,
		)
	}

	return requestBody
}

func (RequestInputReader) MultipartFilesProcessor(
	filesByKey map[string][]*multipart.FileHeader,
) map[string]*multipart.FileHeader {
	fileHeaders := map[string]*multipart.FileHeader{}

	for fileKey, fileHandlers := range filesByKey {
		if len(fileHandlers) == 1 {
			fileHeaders[fileKey] = fileHandlers[0]
			continue
		}

		for fileIndex, fileHandler := range fileHandlers {
			indexedKey := fileKey + "_" + strconv.Itoa(fileIndex)
			fileHeaders[indexedKey] = fileHandler
		}
	}

	return fileHeaders
}

func (reader RequestInputReader) Reader(echoContext echo.Context) (map[string]any, error) {
	requestBody := map[string]any{}

	contentType := echoContext.Request().Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		if err := echoContext.Bind(&requestBody); err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "InvalidJsonBody")
		}

	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		formData, err := echoContext.FormParams()
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "InvalidFormData")
		}
		requestBody = reader.FormUrlEncodedDataProcessor(requestBody, formData)

	case strings.HasPrefix(contentType, "multipart/form-data"):
		multipartForm, err := echoContext.MultipartForm()
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "InvalidMultipartFormData")
		}

		for formKey, formValues := range multipartForm.Value {
			if len(formValues) == 1 {
				requestBody[formKey] = formValues[0]
			}
		}

		if len(multipartForm.File) > 0 {
			fileHeaders := reader.MultipartFilesProcessor(multipartForm.File)
			requestBody["files"] = fileHeaders
		}

	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "InvalidContentType")
	}

	for queryParamName, queryParamValues := range echoContext.QueryParams() {
		requestBody[queryParamName] = queryParamValues[0]
	}

	for _, paramName := range echoContext.ParamNames() {
		requestBody[paramName] = echoContext.Param(paramName)
	}

	if echoContext.Get("operatorAccountId") != nil {
		requestBody["operatorAccountId"] = echoContext.Get("operatorAccountId")
	}
	requestBody["operatorIpAddress"] = echoContext.RealIP()

	return requestBody, nil
}
