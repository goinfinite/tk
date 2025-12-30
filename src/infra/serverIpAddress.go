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
	primaryResolverHostname, err := tkValueObject.NewUnixHostname("whoami.ds.akahelp.net")
	if err != nil {
		return ipAddress, err
	}
	primaryResolver := NewDnsLookup(primaryResolverHostname, &tkValueObject.DnsRecordTypeTXT)
	primaryResolverResults, primaryResolverError := primaryResolver.Execute()
	if primaryResolverError == nil && len(primaryResolverResults) > 0 {
		// ExpectedFormat: "ip1.2.3.4"
		for _, rawRecord := range primaryResolverResults {
			if !strings.Contains(rawRecord, "ip") {
				continue
			}

			rawRecord = strings.TrimPrefix(rawRecord, "ip")
			ipAddress, err = tkValueObject.NewIpAddress(rawRecord)
			if err == nil {
				return ipAddress, nil
			}
		}
	}

	secondaryResolver := NewShell(ShellSettings{
		Command:              "curl",
		Args:                 []string{"-sL", "https://goinfinite.net/ip"},
		ExecutionTimeoutSecs: 3,
	})
	secondaryResolverResults, secondaryResolverError := secondaryResolver.Run()
	if secondaryResolverError == nil && len(secondaryResolverResults) > 0 {
		return tkValueObject.NewIpAddress(secondaryResolverResults)
	}

	return ipAddress, errors.New("PublicIpAddressNotFound")
}
