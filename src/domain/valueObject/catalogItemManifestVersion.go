package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var catalogItemManifestVersionRegex = regexp.MustCompile(`^v\d+(\.\d+)*$`)

type CatalogItemManifestVersion string

func NewCatalogItemManifestVersion(value any) (
	catalogItemManifestVersion CatalogItemManifestVersion, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return catalogItemManifestVersion, errors.New(
			"CatalogItemManifestVersionMustBeString",
		)
	}
	stringValue = strings.ToLower(stringValue)

	if len(stringValue) > 64 {
		return catalogItemManifestVersion, errors.New(
			"CatalogItemManifestVersionTooBig",
		)
	}

	if !catalogItemManifestVersionRegex.MatchString(stringValue) {
		return catalogItemManifestVersion, errors.New(
			"InvalidCatalogItemManifestVersion",
		)
	}

	return CatalogItemManifestVersion(stringValue), nil
}

func (vo CatalogItemManifestVersion) String() string {
	return string(vo)
}
