package tkInfra

import (
	"context"
	"net"
	"strings"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	dnsLookupQueryTimeoutSecsDefault uint = 5
	dnsLookupDialTimeoutMsDefault    uint = 200
)

var (
	dnsLookupPrimaryResolverDefault   = tkValueObject.IpAddress("8.8.8.8")
	dnsLookupSecondaryResolverDefault = tkValueObject.IpAddress("185.228.168.168")
)

type DnsLookupSettings struct {
	PrimaryResolver   tkValueObject.IpAddress
	SecondaryResolver tkValueObject.IpAddress
	QueryTimeoutSecs  uint
	DialTimeoutMs     uint
}

type DnsLookup struct {
	primaryResolver   tkValueObject.IpAddress
	secondaryResolver tkValueObject.IpAddress
	queryTimeoutSecs  uint
	dialTimeoutMs     uint
}

func NewDnsLookup(settings DnsLookupSettings) *DnsLookup {
	primaryResolver := dnsLookupPrimaryResolverDefault
	if settings.PrimaryResolver != "" {
		primaryResolver = settings.PrimaryResolver
	}

	secondaryResolver := dnsLookupSecondaryResolverDefault
	if settings.SecondaryResolver != "" {
		secondaryResolver = settings.SecondaryResolver
	}

	queryTimeoutSecs := dnsLookupQueryTimeoutSecsDefault
	if settings.QueryTimeoutSecs != 0 {
		queryTimeoutSecs = settings.QueryTimeoutSecs
	}

	dialTimeoutMs := dnsLookupDialTimeoutMsDefault
	if settings.DialTimeoutMs != 0 {
		dialTimeoutMs = settings.DialTimeoutMs
	}

	return &DnsLookup{
		primaryResolver:   primaryResolver,
		secondaryResolver: secondaryResolver,
		queryTimeoutSecs:  queryTimeoutSecs,
		dialTimeoutMs:     dialTimeoutMs,
	}
}

func (lookup *DnsLookup) resolverFactory(
	resolverIpAddress tkValueObject.IpAddress,
) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{
				Timeout: time.Duration(lookup.dialTimeoutMs) * time.Millisecond,
			}
			return dialer.DialContext(ctx, "udp", resolverIpAddress.String()+":53")
		},
	}
}

func (lookup *DnsLookup) queryDnsRecords(
	dnsContext context.Context,
	dnsResolver *net.Resolver,
	hostname tkValueObject.UnixHostname,
	recordType tkValueObject.DnsRecordType,
) (queryResults []string, queryError error) {
	hostnameStr := hostname.String()
	switch recordType {
	case tkValueObject.DnsRecordTypeA:
		queryResults, queryError = dnsResolver.LookupHost(dnsContext, hostnameStr)
		var ipv4Addresses []string
		for _, dnsRecord := range queryResults {
			if net.ParseIP(dnsRecord).To4() != nil {
				ipv4Addresses = append(ipv4Addresses, dnsRecord)
			}
		}
		queryResults = ipv4Addresses
	case tkValueObject.DnsRecordTypeAAAA:
		queryResults, queryError = dnsResolver.LookupHost(dnsContext, hostnameStr)
		var ipv6Addresses []string
		for _, dnsRecord := range queryResults {
			if parsedIp := net.ParseIP(dnsRecord); parsedIp != nil && parsedIp.To4() == nil {
				ipv6Addresses = append(ipv6Addresses, dnsRecord)
			}
		}
		queryResults = ipv6Addresses
	case tkValueObject.DnsRecordTypeMX:
		mxRecords, err := dnsResolver.LookupMX(dnsContext, hostnameStr)
		if err != nil {
			return nil, err
		}
		for _, mxRecord := range mxRecords {
			queryResults = append(queryResults, mxRecord.Host)
		}
		queryError = err
	case tkValueObject.DnsRecordTypeTXT:
		queryResults, queryError = dnsResolver.LookupTXT(dnsContext, hostnameStr)
	case tkValueObject.DnsRecordTypeNS:
		nsRecords, err := dnsResolver.LookupNS(dnsContext, hostnameStr)
		if err != nil {
			return nil, err
		}
		for _, nsRecord := range nsRecords {
			queryResults = append(queryResults, nsRecord.Host)
		}
		queryError = err
	case tkValueObject.DnsRecordTypeCNAME:
		cnameRecord, err := dnsResolver.LookupCNAME(dnsContext, hostnameStr)
		if err != nil {
			return nil, err
		}
		queryResults = []string{cnameRecord}
		queryError = err
	case tkValueObject.DnsRecordTypePTR:
		ptrRecords, err := dnsResolver.LookupAddr(dnsContext, hostnameStr)
		if err != nil {
			return nil, err
		}
		queryResults = ptrRecords
		queryError = err
	default:
		queryResults, queryError = dnsResolver.LookupHost(dnsContext, hostnameStr)
	}

	var trimmedResults []string
	for _, rawValue := range queryResults {
		trimmedValue := strings.TrimSpace(rawValue)
		if trimmedValue != "" {
			trimmedResults = append(trimmedResults, trimmedValue)
		}
	}

	return trimmedResults, queryError
}

func (lookup *DnsLookup) Execute(
	hostname tkValueObject.UnixHostname,
	recordType *tkValueObject.DnsRecordType,
) ([]string, error) {
	dnsRecordType := tkValueObject.DnsRecordTypeDefault
	if recordType != nil {
		dnsRecordType = *recordType
	}

	lookupContext, contextCancel := context.WithTimeout(
		context.Background(),
		time.Duration(lookup.queryTimeoutSecs)*time.Second,
	)
	defer contextCancel()

	primaryResolver := lookup.resolverFactory(lookup.primaryResolver)
	zoneRecords, err := lookup.queryDnsRecords(
		lookupContext, primaryResolver, hostname, dnsRecordType,
	)
	if err == nil && len(zoneRecords) > 0 {
		return zoneRecords, nil
	}

	secondaryResolver := lookup.resolverFactory(lookup.secondaryResolver)
	zoneRecords, err = lookup.queryDnsRecords(
		lookupContext, secondaryResolver, hostname, dnsRecordType,
	)
	if err == nil && len(zoneRecords) > 0 {
		return zoneRecords, nil
	}

	return zoneRecords, err
}
