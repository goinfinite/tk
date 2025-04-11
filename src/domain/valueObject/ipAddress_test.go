package tkValueObject

import (
	"testing"
)

var (
	IpAddressesValid   = []string{"192.168.1.1", "10.0.0.1", "172.16.0.1", "::1", "2001:db8::1"}
	IpAddressesInvalid = []string{"192.168.1.256", "300.0.0.1", "123.456.78.90", "abcd::12345"}
)

func TestNewIpAddress(t *testing.T) {
	t.Run("ValidIpAddress", func(t *testing.T) {
		for _, ipAddress := range IpAddressesValid {
			_, err := NewIpAddress(ipAddress)
			if err != nil {
				t.Errorf("ExpectingNoErrorButGot: %s [%s]", err.Error(), ipAddress)
			}
		}
	})

	t.Run("InvalidIpAddress", func(t *testing.T) {
		for _, ipAddress := range IpAddressesInvalid {
			_, err := NewIpAddress(ipAddress)
			if err == nil {
				t.Errorf("ExpectingErrorButGotNil: %s", ipAddress)
			}
		}
	})
}
