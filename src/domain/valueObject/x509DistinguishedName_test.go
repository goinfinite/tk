package tkValueObject

import "testing"

func TestNewX509DistinguishedName(t *testing.T) {
	t.Run("ValidDistinguishedName", func(t *testing.T) {
		org, _ := NewX509Organization("Example Corp")
		ou1, _ := NewX509OrganizationalUnit("Engineering")
		ou2, _ := NewX509OrganizationalUnit("Security")
		locality, _ := NewX509Locality("San Francisco")
		state, _ := NewX509StateOrProvince("California")
		country, _ := NewCountryCode("US")

		testCaseStructs := []struct {
			organization       *X509Organization
			organizationalUnit []X509OrganizationalUnit
			locality           *X509Locality
			stateOrProvince    *X509StateOrProvince
			country            *CountryCode
		}{
			{&org, []X509OrganizationalUnit{ou1}, &locality, &state, &country},
			{&org, []X509OrganizationalUnit{ou1, ou2}, &locality, &state, &country},
			{&org, nil, &locality, &state, &country},
			{&org, nil, nil, nil, nil},
			{nil, nil, nil, nil, nil},
			{nil, []X509OrganizationalUnit{ou1}, &locality, &state, &country},
		}

		for _, testCase := range testCaseStructs {
			result := NewX509DistinguishedName(
				testCase.organization, testCase.organizationalUnit,
				testCase.locality, testCase.stateOrProvince, testCase.country,
			)

			if testCase.organization != nil &&
				result.Organization != testCase.organization {
				t.Errorf("UnexpectedOrganization")
			}

			if testCase.organizationalUnit != nil &&
				len(result.OrganizationalUnit) != len(testCase.organizationalUnit) {
				t.Errorf("UnexpectedOrganizationalUnitLength")
			}

			if testCase.locality != nil && result.Locality != testCase.locality {
				t.Errorf("UnexpectedLocality")
			}

			if testCase.stateOrProvince != nil &&
				result.StateOrProvince != testCase.stateOrProvince {
				t.Errorf("UnexpectedStateOrProvince")
			}

			if testCase.country != nil && result.Country != testCase.country {
				t.Errorf("UnexpectedCountry")
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		org, _ := NewX509Organization("Example Corp")
		ou1, _ := NewX509OrganizationalUnit("Engineering")
		ou2, _ := NewX509OrganizationalUnit("Security")
		locality, _ := NewX509Locality("San Francisco")
		state, _ := NewX509StateOrProvince("California")
		country, _ := NewCountryCode("US")

		testCaseStructs := []struct {
			dn             X509DistinguishedName
			expectedOutput string
		}{
			{
				NewX509DistinguishedName(
					&org, []X509OrganizationalUnit{ou1}, &locality, &state,
					&country,
				),
				"O=Example Corp, OU=Engineering, L=San Francisco, ST=California, C=US",
			},
			{
				NewX509DistinguishedName(
					&org, []X509OrganizationalUnit{ou1, ou2}, &locality,
					&state, &country,
				),
				"O=Example Corp, OU=Engineering, OU=Security, L=San Francisco, ST=California, C=US",
			},
			{
				NewX509DistinguishedName(&org, nil, nil, nil, nil),
				"O=Example Corp",
			},
			{
				NewX509DistinguishedName(nil, nil, nil, nil, nil),
				"",
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.dn.String()

			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v'", actualOutput, testCase.expectedOutput)
			}
		}
	})
}
