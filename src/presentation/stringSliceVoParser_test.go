package tkPresentation

import (
	"strings"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestStringSliceValueObjectParser(t *testing.T) {
	var (
		IpAddressesValid   = []string{"192.168.1.1", "10.0.0.1", "172.16.0.1", "::1", "2001:db8::1"}
		IpAddressesInvalid = []string{"192.168.1.256", "300.0.0.1", "123.456.78.90", "abcd::12345"}
	)

	t.Run("NilInput", func(t *testing.T) {
		parsedObjects := StringSliceValueObjectParser(nil, tkValueObject.NewIpAddress)
		if len(parsedObjects) != 0 {
			t.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})

	t.Run("InvalidStringValueInput", func(t *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesInvalid[0], tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 0 {
			t.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})

	t.Run("SingleStringValueInput", func(t *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesValid[0], tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 1 {
			t.Errorf("ExpectedSliceWith1ElementButGot: %v", parsedObjects)
		}
	})

	t.Run("MultipleStringInput", func(t *testing.T) {
		validIpAddressesSeparatedBySemicolon := strings.Join(
			IpAddressesValid, ";",
		)

		parsedObjects := StringSliceValueObjectParser(
			validIpAddressesSeparatedBySemicolon, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 5 {
			t.Errorf("ExpectedSliceWith5ElementsButGot: %v", parsedObjects)
		}

		validIpAddressesSeparatedByComma := strings.Join(
			IpAddressesValid, ",",
		)

		parsedObjects = StringSliceValueObjectParser(
			validIpAddressesSeparatedByComma, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 5 {
			t.Errorf("ExpectedSliceWith5ElementsButGot: %v", parsedObjects)
		}
	})

	t.Run("InvalidMultipleStringInput", func(t *testing.T) {
		invalidIpAddressesSeparatedBySemicolon := strings.Join(
			IpAddressesInvalid, ";",
		)

		parsedObjects := StringSliceValueObjectParser(
			invalidIpAddressesSeparatedBySemicolon, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 0 {
			t.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})

	t.Run("SliceInput", func(t *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesValid, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 5 {
			t.Errorf("ExpectedSliceWith5ElementsButGot: %v", parsedObjects)
		}
	})

	t.Run("InvalidSliceInput", func(t *testing.T) {
		parsedObjects := StringSliceValueObjectParser(
			IpAddressesInvalid, tkValueObject.NewIpAddress,
		)
		if len(parsedObjects) != 0 {
			t.Errorf("ExpectedEmptySliceButGot: %v", parsedObjects)
		}
	})
}
