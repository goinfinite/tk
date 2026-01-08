package tkValueObject

import (
	"crypto/x509/pkix"
	"log/slog"
	"strings"
)

type X509DistinguishedName struct {
	Organization       *X509Organization        `json:"organization"`
	OrganizationalUnit []X509OrganizationalUnit `json:"organizationalUnit"`
	Locality           *X509Locality            `json:"locality"`
	StateOrProvince    *X509StateOrProvince     `json:"stateOrProvince"`
	Country            *CountryCode             `json:"country"`
}

func NewX509DistinguishedName(
	organization *X509Organization,
	organizationalUnit []X509OrganizationalUnit,
	locality *X509Locality,
	stateOrProvince *X509StateOrProvince,
	country *CountryCode,
) X509DistinguishedName {
	return X509DistinguishedName{
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		StateOrProvince:    stateOrProvince,
		Country:            country,
	}
}

func NewX509DistinguishedNameFromPkixName(
	pkixName pkix.Name,
) (*X509DistinguishedName, error) {
	var organizationPtr *X509Organization
	if len(pkixName.Organization) > 0 {
		organization, err := NewX509Organization(pkixName.Organization[0])
		if err != nil {
			return nil, err
		}
		organizationPtr = &organization
	}

	var organizationalUnits []X509OrganizationalUnit
	for _, rawOrganizationalUnit := range pkixName.OrganizationalUnit {
		organizationalUnit, err := NewX509OrganizationalUnit(
			rawOrganizationalUnit,
		)
		if err != nil {
			slog.Debug(
				"SkipInvalidOrganizationalUnit",
				slog.String("value", rawOrganizationalUnit),
			)
			continue
		}
		organizationalUnits = append(organizationalUnits, organizationalUnit)
	}

	var localityPtr *X509Locality
	if len(pkixName.Locality) > 0 {
		locality, err := NewX509Locality(pkixName.Locality[0])
		if err != nil {
			return nil, err
		}
		localityPtr = &locality
	}

	var stateOrProvincePtr *X509StateOrProvince
	if len(pkixName.Province) > 0 {
		stateOrProvince, err := NewX509StateOrProvince(pkixName.Province[0])
		if err != nil {
			return nil, err
		}
		stateOrProvincePtr = &stateOrProvince
	}

	var countryPtr *CountryCode
	if len(pkixName.Country) > 0 {
		country, err := NewCountryCode(pkixName.Country[0])
		if err != nil {
			return nil, err
		}
		countryPtr = &country
	}

	distinguishedName := NewX509DistinguishedName(
		organizationPtr, organizationalUnits, localityPtr,
		stateOrProvincePtr, countryPtr,
	)
	return &distinguishedName, nil
}

func (vo X509DistinguishedName) String() string {
	var builder strings.Builder
	needsSeparator := false

	if vo.Organization != nil {
		builder.WriteString("O=")
		builder.WriteString(vo.Organization.String())
		needsSeparator = true
	}

	for _, ou := range vo.OrganizationalUnit {
		if needsSeparator {
			builder.WriteString(", ")
		}
		builder.WriteString("OU=")
		builder.WriteString(ou.String())
		needsSeparator = true
	}

	if vo.Locality != nil {
		if needsSeparator {
			builder.WriteString(", ")
		}
		builder.WriteString("L=")
		builder.WriteString(vo.Locality.String())
		needsSeparator = true
	}

	if vo.StateOrProvince != nil {
		if needsSeparator {
			builder.WriteString(", ")
		}
		builder.WriteString("ST=")
		builder.WriteString(vo.StateOrProvince.String())
		needsSeparator = true
	}

	if vo.Country != nil {
		if needsSeparator {
			builder.WriteString(", ")
		}
		builder.WriteString("C=")
		builder.WriteString(vo.Country.String())
	}

	return builder.String()
}
