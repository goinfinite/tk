package tkInfra

import (
	"context"
	"net"
	"strings"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type DnsLookup struct {
	primaryResolver   tkValueObject.IpAddress
	secondaryResolver tkValueObject.IpAddress
	queryTimeoutSec   uint
	dialTimeoutMs     uint
	hostname          tkValueObject.UnixHostname
	recordType        tkValueObject.DnsRecordType
}

func NewDnsLookup(
	hostname tkValueObject.UnixHostname,
	recordType *tkValueObject.DnsRecordType,
) *DnsLookup {
	dnsRecordType := tkValueObject.DnsRecordTypeDefault
	if recordType != nil {
		dnsRecordType = *recordType
	}

	return &DnsLookup{
		primaryResolver:   tkValueObject.IpAddress("8.8.8.8"),
		secondaryResolver: tkValueObject.IpAddress("185.228.168.168"),
		queryTimeoutSec:   5,
		dialTimeoutMs:     200,
		hostname:          hostname,
		recordType:        dnsRecordType,
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
) (queryResults []string, queryError error) {
	switch lookup.recordType {
	case tkValueObject.DnsRecordTypeA:
		queryResults, queryError = dnsResolver.LookupHost(dnsContext, lookup.hostname.String())
		var ipv4Addresses []string
		for _, dnsRecord := range queryResults {
			if net.ParseIP(dnsRecord).To4() != nil {
				ipv4Addresses = append(ipv4Addresses, dnsRecord)
			}
		}
		queryResults = ipv4Addresses
	case tkValueObject.DnsRecordTypeAAAA:
		queryResults, queryError = dnsResolver.LookupHost(dnsContext, lookup.hostname.String())
		var ipv6Addresses []string
		for _, dnsRecord := range queryResults {
			if parsedIp := net.ParseIP(dnsRecord); parsedIp != nil && parsedIp.To4() == nil {
				ipv6Addresses = append(ipv6Addresses, dnsRecord)
			}
		}
		queryResults = ipv6Addresses
	case tkValueObject.DnsRecordTypeMX:
		mxRecords, err := dnsResolver.LookupMX(dnsContext, lookup.hostname.String())
		if err != nil {
			return nil, err
		}
		for _, mxRecord := range mxRecords {
			queryResults = append(queryResults, mxRecord.Host)
		}
		queryError = err
	case tkValueObject.DnsRecordTypeTXT:
		queryResults, queryError = dnsResolver.LookupTXT(dnsContext, lookup.hostname.String())
	case tkValueObject.DnsRecordTypeNS:
		nsRecords, err := dnsResolver.LookupNS(dnsContext, lookup.hostname.String())
		if err != nil {
			return nil, err
		}
		for _, nsRecord := range nsRecords {
			queryResults = append(queryResults, nsRecord.Host)
		}
		queryError = err
	case tkValueObject.DnsRecordTypeCNAME:
		cnameRecord, err := dnsResolver.LookupCNAME(dnsContext, lookup.hostname.String())
		if err != nil {
			return nil, err
		}
		queryResults = []string{cnameRecord}
		queryError = err
	case tkValueObject.DnsRecordTypePTR:
		ptrRecords, err := dnsResolver.LookupAddr(dnsContext, lookup.hostname.String())
		if err != nil {
			return nil, err
		}
		queryResults = ptrRecords
		queryError = err
	default:
		queryResults, queryError = dnsResolver.LookupHost(dnsContext, lookup.hostname.String())
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

func (lookup *DnsLookup) Execute() ([]string, error) {
	lookupContext, contextCancel := context.WithTimeout(
		context.Background(), time.Duration(lookup.queryTimeoutSec)*time.Second,
	)
	defer contextCancel()

	primaryResolver := lookup.resolverFactory(lookup.primaryResolver)
	zoneRecords, err := lookup.queryDnsRecords(lookupContext, primaryResolver)
	if err == nil && len(zoneRecords) > 0 {
		return zoneRecords, nil
	}

	secondaryResolver := lookup.resolverFactory(lookup.secondaryResolver)
	zoneRecords, err = lookup.queryDnsRecords(lookupContext, secondaryResolver)
	if err == nil && len(zoneRecords) > 0 {
		return zoneRecords, nil
	}

	return zoneRecords, err
}
