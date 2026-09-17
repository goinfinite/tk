package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var catalogItemEnvRegex = regexp.MustCompile(`^\w{1,1000}=.{1,1000}$`)

type CatalogItemEnv string

func NewCatalogItemEnv(value any) (catalogItemEnv CatalogItemEnv, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return catalogItemEnv, errors.New("CatalogItemEnvMustBeString")
	}

	if !catalogItemEnvRegex.MatchString(stringValue) {
		return catalogItemEnv, errors.New("InvalidCatalogItemEnv")
	}

	return CatalogItemEnv(stringValue), nil
}

func (vo CatalogItemEnv) ReadKey() string {
	key, _, _ := strings.Cut(string(vo), "=")
	return key
}

func (vo CatalogItemEnv) ReadValue() string {
	_, value, _ := strings.Cut(string(vo), "=")
	return value
}

func (vo CatalogItemEnv) String() string {
	return string(vo)
}
