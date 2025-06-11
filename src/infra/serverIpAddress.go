package tkInfra

import (
	"errors"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func ReadServerPrivateIpAddress() (ipAddress tkValueObject.IpAddress, err error) {
	rawIpAddress, err := NewShell(
		ShellSettings{Command: "hostname", Args: []string{"-I"}},
	).Run()
	if err != nil {
		return ipAddress, err
	}

	rawIpAddresses := strings.Split(rawIpAddress, " ")
	if len(rawIpAddresses) > 0 {
		rawIpAddress = rawIpAddresses[0]
	}

	rawIpAddress = strings.TrimSpace(rawIpAddress)
	if rawIpAddress == "" {
		return ipAddress, errors.New("PrivateIpAddressNotFound")
	}

	return tkValueObject.NewIpAddress(rawIpAddress)
}

func ReadServerPublicIpAddress() (ipAddress tkValueObject.IpAddress, err error) {
	rawIpAddress, err := NewShell(ShellSettings{
		Command: "dig", Args: []string{"+short", "TXT", "o-o.myaddr.l.google.com", "@ns1.google.com"},
	}).Run()
	if err != nil || rawIpAddress == "" {
		rawIpAddress, err = NewShell(ShellSettings{
			Command: "dig", Args: []string{"+short", "TXT", "CH", "whoami.cloudflare", "@1.1.1.1"},
		}).Run()
		if err != nil {
			return ipAddress, err
		}
	}

	rawIpAddress = strings.Trim(rawIpAddress, `"`)
	rawIpAddress = strings.TrimSpace(rawIpAddress)
	if rawIpAddress == "" {
		return ipAddress, errors.New("PublicIpAddressNotFound")
	}

	ipAddress, err = tkValueObject.NewIpAddress(rawIpAddress)
	if err != nil {
		return ipAddress, err
	}

	return ipAddress, nil
}
