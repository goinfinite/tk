package tkUseCase

import (
	"log/slog"
	"strings"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type activityRecordLogAttributesInput struct {
	Pagination        *tkDto.Pagination
	RecordId          *tkValueObject.ActivityRecordId
	RecordLevel       *tkValueObject.ActivityRecordLevel
	RecordCode        *tkValueObject.ActivityRecordCode
	AffectedResources []tkValueObject.SystemResourceIdentifier
	OperatorSri       *tkValueObject.SystemResourceIdentifier
	OperatorIpAddress *tkValueObject.IpAddress
	CreatedBeforeAt   *tkValueObject.UnixTime
	CreatedAfterAt    *tkValueObject.UnixTime
}

func activityRecordLogAttributesBuilder(
	input activityRecordLogAttributesInput,
) []any {
	logAttributes := []any{
		slog.Int("affectedResourcesCount", len(input.AffectedResources)),
	}

	if len(input.AffectedResources) > 0 {
		resourceIdentifiers := make([]string, 0, len(input.AffectedResources))
		for _, affectedResource := range input.AffectedResources {
			resourceIdentifiers = append(
				resourceIdentifiers, affectedResource.String(),
			)
		}
		logAttributes = append(
			logAttributes,
			slog.String(
				"affectedResources", strings.Join(resourceIdentifiers, ","),
			),
		)
	}

	if input.Pagination != nil {
		logAttributes = append(
			logAttributes,
			slog.Int("pageNumber", int(input.Pagination.PageNumber)),
			slog.Int("itemsPerPage", int(input.Pagination.ItemsPerPage)),
		)
		if input.Pagination.SortBy != nil {
			logAttributes = append(
				logAttributes,
				slog.String("sortBy", input.Pagination.SortBy.String()),
			)
		}
		if input.Pagination.SortDirection != nil {
			logAttributes = append(
				logAttributes,
				slog.String(
					"sortDirection",
					input.Pagination.SortDirection.String(),
				),
			)
		}
		if input.Pagination.LastSeenId != nil {
			logAttributes = append(
				logAttributes,
				slog.String("lastSeenId", input.Pagination.LastSeenId.String()),
			)
		}
	}

	if input.RecordId != nil {
		logAttributes = append(
			logAttributes, slog.String("recordId", input.RecordId.String()),
		)
	}
	if input.RecordLevel != nil {
		logAttributes = append(
			logAttributes,
			slog.String("recordLevel", input.RecordLevel.String()),
		)
	}
	if input.RecordCode != nil {
		logAttributes = append(
			logAttributes,
			slog.String("recordCode", input.RecordCode.String()),
		)
	}
	if input.OperatorSri != nil {
		logAttributes = append(
			logAttributes,
			slog.String("operatorSri", input.OperatorSri.String()),
		)
	}
	if input.OperatorIpAddress != nil {
		logAttributes = append(
			logAttributes,
			slog.String("operatorIpAddress", input.OperatorIpAddress.String()),
		)
	}
	if input.CreatedBeforeAt != nil {
		logAttributes = append(
			logAttributes,
			slog.String("createdBeforeAt", input.CreatedBeforeAt.String()),
		)
	}
	if input.CreatedAfterAt != nil {
		logAttributes = append(
			logAttributes,
			slog.String("createdAfterAt", input.CreatedAfterAt.String()),
		)
	}

	return logAttributes
}
