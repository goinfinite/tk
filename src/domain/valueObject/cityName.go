package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var cityNameRegex = regexp.MustCompile(`^\p{L}[\p{L}\'\ \-]{2,128}$`)

type CityName string

func NewCityName(value any) (cityName CityName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return cityName, errors.New("CityNameMustBeString")
	}

	capitalizedCityName := cases.Title(language.English, cases.Compact).String(stringValue)

	if !cityNameRegex.MatchString(capitalizedCityName) {
		return cityName, errors.New("InvalidCityName")
	}

	return CityName(capitalizedCityName), nil
}

func (vo CityName) String() string {
	return string(vo)
}
