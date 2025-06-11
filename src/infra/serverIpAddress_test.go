package tkInfra

import (
	"testing"
)

func TestServerIpAddress(t *testing.T) {
	t.Run("FunctionalTest", func(t *testing.T) {
		_, err := ReadServerPrivateIpAddress()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		_, err = ReadServerPublicIpAddress()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}
	})
}
