package tkPresentation

import (
	"strings"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestStringSliceValueObjectParser(test *testing.T) {
	var (
		IpAddressesValid   = []string{"192.168.1.1", "10.0.0.1", "172.16.0.1", "::1", "2001:db8::1"}
		IpAddressesInvalid = []string{"192.168.1.256", "300.0.0.1", "123.456.78.90", "abcd::12345"}
	)

	test.Run("NilInput", func(subtest *testing.T) {
		parsedObjects := StringSliceValueObjectParser(nil, tkValueObject.NewIpAddress)
		if len(parsedObjects) != 0 {
			subtest.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})

	test.Run("InvalidStringValueInput", func(subtest *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesInvalid[0], tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 0 {
			subtest.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})

	test.Run("SingleStringValueInput", func(subtest *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesValid[0], tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 1 {
			subtest.Errorf("ExpectedSliceWith1ElementButGot: %v", parsedObjects)
		}
	})

	test.Run("MultipleStringInput", func(subtest *testing.T) {
		validIpAddressesSeparatedBySemicolon := strings.Join(
			IpAddressesValid, ";",
		)

		parsedObjects := StringSliceValueObjectParser(
			validIpAddressesSeparatedBySemicolon, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 5 {
			subtest.Errorf("ExpectedSliceWith5ElementsButGot: %v", parsedObjects)
		}

		validIpAddressesSeparatedByComma := strings.Join(
			IpAddressesValid, ",",
		)

		parsedObjects = StringSliceValueObjectParser(
			validIpAddressesSeparatedByComma, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 5 {
			subtest.Errorf("ExpectedSliceWith5ElementsButGot: %v", parsedObjects)
		}
	})

	test.Run("InvalidMultipleStringInput", func(subtest *testing.T) {
		invalidIpAddressesSeparatedBySemicolon := strings.Join(
			IpAddressesInvalid, ";",
		)

		parsedObjects := StringSliceValueObjectParser(
			invalidIpAddressesSeparatedBySemicolon, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 0 {
			subtest.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})

	test.Run("SliceInput", func(subtest *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesValid, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 5 {
			subtest.Errorf("ExpectedSliceWith5ElementsButGot: %v", parsedObjects)
		}
	})

	test.Run("InvalidSliceInput", func(subtest *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesInvalid, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 0 {
			subtest.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})
}
