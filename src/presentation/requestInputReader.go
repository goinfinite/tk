package tkPresentation

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

type ApiRequestInputReader struct {
}

func (reader ApiRequestInputReader) StringDotNotationToHierarchicalMap(
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

	parentHierarchicalMap, assertOk := hierarchicalMap[parentKey].(map[string]any)
	if !assertOk {
		return hierarchicalMap
	}

	hierarchicalMap[parentKey] = reader.StringDotNotationToHierarchicalMap(
		parentHierarchicalMap, nextKeys, finalValue,
	)

	return hierarchicalMap
}

func (reader ApiRequestInputReader) FormUrlEncodedDataProcessor(
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
		requestBody = reader.StringDotNotationToHierarchicalMap(
			requestBody, keyParts, formValue,
		)
	}

	return requestBody
}

func (ApiRequestInputReader) MultipartFilesProcessor(
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

// ApiRequestInputReader.Reader extracts and normalizes input data from an HTTP request
// into a flat map[string]any.
//
// It processes the request body based on Content-Type, merges query parameters,
// path parameters, and optionally includes operator context information.
//
// Content-Type handling:
//   - application/json: Parses JSON body directly into the map. Returns "InvalidJsonBody"
//     error if JSON is malformed.
//   - application/x-www-form-urlencoded: Processes form data, preserving multiple values
//     as string slices when a key appears more than once. Supports dot notation for nested
//     structures (e.g., "user.name" becomes map["user"]["name"]).
//   - multipart/form-data: Processes form fields similar to URL-encoded data. File uploads
//     are normalized under the "files" key, with multiple files indexed as "key_0", "key_1", etc.
//     Returns "InvalidMultipartFormData" error if the multipart form cannot be parsed.
//   - Other/missing Content-Type: Returns "InvalidContentType" error.
//
// Query parameters:
//   - Query parameters are merged into the result map. When a query parameter has multiple
//     values (e.g., ?tags=tag1&tags=tag2), only the FIRST value is captured. This is a design
//     decision to optimize for the common single-value case (~99% of use cases).
//   - For the rare cases where multiple values are needed, use a delimiter at the controller
//     level. For example: ?tags=tag1;tag2 and split on ";" in your controller logic.
//
// Route Path parameters:
//   - All Echo route path parameters (e.g., :id, :name) are merged into the result map.
//
// Operator context (if present in Echo context):
//   - operatorSri: Extracted from context key and included in the result map.
//   - operatorIpAddress: If missing, populated using the RealIP() method.
//
// Returns:
//   - A map[string]any containing all extracted request data, or
//   - An echo.HTTPError with status 400 (Bad Request) and a descriptive message if parsing fails.
func (reader ApiRequestInputReader) Reader(echoContext echo.Context) (map[string]any, error) {
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

		if multipartForm == nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "InvalidMultipartFormData")
		}
		requestBody = reader.FormUrlEncodedDataProcessor(requestBody, multipartForm.Value)

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

	for _, routeParamName := range echoContext.ParamNames() {
		requestBody[routeParamName] = echoContext.Param(routeParamName)
	}

	// The `operatorSri` and `operatorIpAddress` fields in the request body are
	// typically extracted from an authentication middleware and passed by the
	// controller in the request body if the API relies on a liaison layer to
	// process the request uniformly. To prevent repeating the same logic in every
	// controller and to ensure the untrusted user doesn't succeed in injecting a
	// fake operator context, we populate/overwrite these values here.
	if operatorSri, assertOk := echoContext.Get("operatorSri").(tkValueObject.SystemResourceIdentifier); assertOk {
		requestBody["operatorSri"] = operatorSri
	}

	if operatorIpAddress, assertOk := echoContext.Get("operatorIpAddress").(tkValueObject.IpAddress); assertOk {
		requestBody["operatorIpAddress"] = operatorIpAddress
	}
	if requestBody["operatorIpAddress"] == nil {
		operatorIpAddress, err := tkValueObject.NewIpAddress(echoContext.RealIP())
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "InvalidOperatorIpAddress")
		}
		requestBody["operatorIpAddress"] = operatorIpAddress
	}

	return requestBody, nil
}
