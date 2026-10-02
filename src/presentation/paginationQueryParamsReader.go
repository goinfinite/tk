package tkPresentation

import (
	"log/slog"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
	"github.com/labstack/echo/v4"
)

func queryParamConverter[TargetType any](
	echoContext echo.Context, paramName string,
	converter func(any) (TargetType, error),
) (convertedValue TargetType, isEmpty bool) {
	rawValue := echoContext.QueryParam(paramName)
	if rawValue == "" {
		return convertedValue, true
	}

	convertedValue, conversionErr := converter(rawValue)
	if conversionErr != nil {
		slog.Debug(
			"InvalidQueryParam",
			slog.String("paramName", paramName),
			slog.String("err", conversionErr.Error()),
		)
		return convertedValue, true
	}

	return convertedValue, false
}

func PaginationQueryParamsReader(
	echoContext echo.Context, optionalParamNames []string,
) map[string]any {
	requestParams := map[string]any{}

	pageNumber, isPageNumberEmpty := queryParamConverter(
		echoContext, "pageNumber", tkVoUtil.InterfaceToUint32,
	)
	if !isPageNumberEmpty {
		requestParams["pageNumber"] = pageNumber
	}

	itemsPerPage, isItemsPerPageEmpty := queryParamConverter(
		echoContext, "itemsPerPage", tkVoUtil.InterfaceToUint16,
	)
	if !isItemsPerPageEmpty && itemsPerPage > 0 {
		requestParams["itemsPerPage"] = itemsPerPage
	}

	sortBy, isSortByEmpty := queryParamConverter(
		echoContext, "sortBy", tkValueObject.NewPaginationSortBy,
	)
	if !isSortByEmpty {
		requestParams["sortBy"] = sortBy
	}

	sortDirection, isSortDirectionEmpty := queryParamConverter(
		echoContext, "sortDirection", tkValueObject.NewPaginationSortDirection,
	)
	if !isSortDirectionEmpty {
		requestParams["sortDirection"] = sortDirection
	}

	lastSeenId, isLastSeenIdEmpty := queryParamConverter(
		echoContext, "lastSeenId", tkValueObject.NewPaginationLastSeenId,
	)
	if !isLastSeenIdEmpty {
		requestParams["lastSeenId"] = lastSeenId
	}

	for _, optionalParamName := range optionalParamNames {
		switch optionalParamName {
		case "pageNumber", "itemsPerPage", "sortBy", "sortDirection", "lastSeenId":
			continue
		}

		optionalParamValue := echoContext.QueryParam(optionalParamName)
		if optionalParamValue == "" {
			continue
		}
		requestParams[optionalParamName] = optionalParamValue
	}

	return requestParams
}
