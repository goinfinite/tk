package tkInfra

import (
	"testing"
)

func TestServerIpAddress(t *testing.T) {
	t.Run("FunctionalTest", func(t *testing.T) {
		t.Setenv(ServerPublicIpAddressEnvVarName, "")

		_, err := ReadServerPrivateIpAddress()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		_, err = ReadServerPublicIpAddress()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}
	})

	t.Run("PublicIpAddressEnvVarTakesPriority", func(t *testing.T) {
		rawEnvIpAddress := "203.0.113.42"
		t.Setenv(ServerPublicIpAddressEnvVarName, rawEnvIpAddress)

		ipAddress, err := ReadServerPublicIpAddress()
		if err != nil {
			t.Fatalf("UnexpectedError: '%s'", err.Error())
		}

		if ipAddress.String() != rawEnvIpAddress {
			t.Errorf("ExpectedIpFromEnvVar '%s', got '%s'", rawEnvIpAddress, ipAddress.String())
		}
	})
}

