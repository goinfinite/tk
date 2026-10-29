package tkPresentation

import (
	"errors"
	"log/slog"
	"maps"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

type UiQueryParamFilterKind string

const (
	UiQueryParamFilterKindScalar UiQueryParamFilterKind = "scalar"
	UiQueryParamFilterKindMulti  UiQueryParamFilterKind = "multi"
	UiQueryParamFilterKindRange  UiQueryParamFilterKind = "range"
)

type UiQueryParamFilter struct {
	Key            string
	QueryParamName string
	Kind           UiQueryParamFilterKind
}

type UiRequestInputReader struct {
	Filters   []UiQueryParamFilter
	BaseInput map[string]any
}

func (reader UiRequestInputReader) isReservedInputKey(inputKey string) bool {
	switch strings.ToLower(inputKey) {
	case "pagenumber", "itemsperpage", "sortby", "sortdirection",
		"lastseenid", "operatoraccountid", "operatorsri",
		"operatoripaddress":
		return true
	}
	return false
}

func (reader UiRequestInputReader) filtersReader(
	echoContext echo.Context,
) map[string]any {
	filtersInput := map[string]any{}
	for _, filter := range reader.Filters {
		if filter.Key == "" || reader.isReservedInputKey(filter.Key) {
			continue
		}

		queryParamName := filter.QueryParamName
		if queryParamName == "" {
			queryParamName = filter.Key
		}

		switch filter.Kind {
		case UiQueryParamFilterKindMulti:
			queryParamValues := echoContext.QueryParams()[queryParamName]
			nonEmptyValues := make([]string, 0, len(queryParamValues))
			for _, queryParamValue := range queryParamValues {
				if queryParamValue == "" {
					continue
				}
				nonEmptyValues = append(nonEmptyValues, queryParamValue)
			}
			if len(nonEmptyValues) == 0 {
				continue
			}
			filtersInput[filter.Key] = nonEmptyValues
		case UiQueryParamFilterKindRange:
			minValue := echoContext.QueryParam(queryParamName + "Min")
			maxValue := echoContext.QueryParam(queryParamName + "Max")
			if minValue == "" && maxValue == "" {
				continue
			}
			filtersInput[filter.Key] = map[string]string{
				"min": minValue,
				"max": maxValue,
			}
		default:
			queryParamValue := echoContext.QueryParam(queryParamName)
			if queryParamValue == "" {
				continue
			}
			filtersInput[filter.Key] = queryParamValue
		}
	}

	return filtersInput
}

func (reader UiRequestInputReader) operatorContextReader(
	echoContext echo.Context,
) (map[string]any, error) {
	operatorContext := map[string]any{
		"operatorAccountId": nil,
		"operatorSri":       nil,
		"operatorIpAddress": nil,
	}

	operatorAccountId, isOperatorAccountId := echoContext.Get(
		"operatorAccountId",
	).(tkValueObject.AccountId)
	if !isOperatorAccountId {
		return nil, errors.New("OperatorAccountIdNotFound")
	}
	operatorContext["operatorAccountId"] = operatorAccountId

	if operatorSri, isOperatorSri := echoContext.Get(
		"operatorSri",
	).(tkValueObject.SystemResourceIdentifier); isOperatorSri {
		operatorContext["operatorSri"] = operatorSri
	}

	operatorIpAddress, isOperatorIpAddress := echoContext.Get(
		"operatorIpAddress",
	).(tkValueObject.IpAddress)
	if !isOperatorIpAddress {
		extractedIpAddress, err := NewRequesterIpExtractor().Execute(
			echoContext.Request(),
		)
		if err != nil {
			slog.Debug(
				"UiRequestInputReaderOperatorIpAddressError",
				slog.String("error", err.Error()),
			)
			return nil, errors.New("InvalidOperatorIpAddress")
		}
		operatorIpAddress = extractedIpAddress
	}
	operatorContext["operatorIpAddress"] = operatorIpAddress

	return operatorContext, nil
}

func (reader UiRequestInputReader) Read(
	echoContext echo.Context,
) (map[string]any, error) {
	operatorContext, err := reader.operatorContextReader(echoContext)
	if err != nil {
		return nil, err
	}

	requestInput := map[string]any{}
	maps.Copy(requestInput, reader.BaseInput)
	maps.Copy(requestInput, reader.filtersReader(echoContext))
	maps.Copy(requestInput, PaginationQueryParamsReader(echoContext, nil))
	maps.Copy(requestInput, operatorContext)

	return requestInput, nil
}
