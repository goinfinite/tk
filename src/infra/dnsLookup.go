package tkInfra

import (
	"context"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"

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
	PrimaryResolver           tkValueObject.IpAddress
	SecondaryResolver         tkValueObject.IpAddress
	QueryTimeoutSecs          uint
	DialTimeoutMs             uint
	ShouldBypassLocalResolver bool
}

type DnsLookup struct {
	primaryResolver           tkValueObject.IpAddress
	secondaryResolver         tkValueObject.IpAddress
	queryTimeoutSecs          uint
	dialTimeoutMs             uint
	shouldBypassLocalResolver bool
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
		primaryResolver:           primaryResolver,
		secondaryResolver:         secondaryResolver,
		queryTimeoutSecs:          queryTimeoutSecs,
		dialTimeoutMs:             dialTimeoutMs,
		shouldBypassLocalResolver: settings.ShouldBypassLocalResolver,
	}
}

func (lookup *DnsLookup) resolverBuilder(
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

func dnsQueryBuilder(
	hostname tkValueObject.UnixHostname,
	questionType dnsmessage.Type,
) (queryBytes []byte, buildError error) {
	dnsName, nameError := dnsmessage.NewName(hostname.String() + ".")
	if nameError != nil {
		return nil, nameError
	}

	queryMessage := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:               1,
			RecursionDesired: true,
		},
		Questions: []dnsmessage.Question{{
			Name:  dnsName,
			Type:  questionType,
			Class: dnsmessage.ClassINET,
		}},
	}

	return queryMessage.Pack()
}

func (lookup *DnsLookup) exchangeDnsMessage(
	dnsContext context.Context,
	resolverIpAddress tkValueObject.IpAddress,
	queryBytes []byte,
) (responseBytes []byte, exchangeError error) {
	dialer := net.Dialer{
		Timeout: time.Duration(lookup.dialTimeoutMs) * time.Millisecond,
	}

	udpConn, dialError := dialer.DialContext(
		dnsContext, "udp", resolverIpAddress.String()+":53",
	)
	if dialError != nil {
		return nil, dialError
	}
	defer udpConn.Close()

	deadline := time.Now().Add(time.Duration(lookup.queryTimeoutSecs) * time.Second)
	deadlineError := udpConn.SetDeadline(deadline)
	if deadlineError != nil {
		return nil, deadlineError
	}

	_, writeError := udpConn.Write(queryBytes)
	if writeError != nil {
		return nil, writeError
	}

	rawResponseBuffer := make([]byte, 512)
	bytesRead, readError := udpConn.Read(rawResponseBuffer)
	if readError != nil {
		return nil, readError
	}

	return rawResponseBuffer[:bytesRead], nil
}

func dnsMessageResponseIpAddressExtractor(
	responseBytes []byte,
	questionType dnsmessage.Type,
) (ipAddresses []string, decodeError error) {
	var responseMessage dnsmessage.Message
	unpackError := responseMessage.Unpack(responseBytes)
	if unpackError != nil {
		return nil, unpackError
	}

	for _, answer := range responseMessage.Answers {
		switch typedAnswer := answer.Body.(type) {
		case *dnsmessage.AResource:
			ipAddresses = append(ipAddresses, net.IP(typedAnswer.A[:]).String())
		case *dnsmessage.AAAAResource:
			ipAddresses = append(ipAddresses, net.IP(typedAnswer.AAAA[:]).String())
		}
	}

	return ipAddresses, nil
}

func (lookup *DnsLookup) directIpAddressResolver(
	dnsContext context.Context,
	resolverIpAddress tkValueObject.IpAddress,
	hostname tkValueObject.UnixHostname,
	recordType tkValueObject.DnsRecordType,
) (ipAddresses []string, queryError error) {
	var questionType dnsmessage.Type
	switch recordType {
	case tkValueObject.DnsRecordTypeA:
		questionType = dnsmessage.TypeA
	case tkValueObject.DnsRecordTypeAAAA:
		questionType = dnsmessage.TypeAAAA
	}

	queryBytes, buildError := dnsQueryBuilder(hostname, questionType)
	if buildError != nil {
		return nil, buildError
	}

	responseBytes, exchangeError := lookup.exchangeDnsMessage(
		dnsContext, resolverIpAddress, queryBytes,
	)
	if exchangeError != nil {
		return nil, exchangeError
	}

	return dnsMessageResponseIpAddressExtractor(responseBytes, questionType)
}

func (lookup *DnsLookup) defaultDnsRecordsResolver(
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

func (lookup *DnsLookup) dnsRecordsResolver(
	dnsContext context.Context,
	resolverIpAddress tkValueObject.IpAddress,
	hostname tkValueObject.UnixHostname,
	recordType tkValueObject.DnsRecordType,
) ([]string, error) {
	isIpAddressRecordType := recordType == tkValueObject.DnsRecordTypeA ||
		recordType == tkValueObject.DnsRecordTypeAAAA
	if lookup.shouldBypassLocalResolver && isIpAddressRecordType {
		return lookup.directIpAddressResolver(
			dnsContext, resolverIpAddress, hostname, recordType,
		)
	}

	resolver := lookup.resolverBuilder(resolverIpAddress)
	return lookup.defaultDnsRecordsResolver(
		dnsContext, resolver, hostname, recordType,
	)
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

	primaryResults, err := lookup.dnsRecordsResolver(
		lookupContext, lookup.primaryResolver, hostname, dnsRecordType,
	)
	if err == nil && len(primaryResults) > 0 {
		return primaryResults, nil
	}

	secondaryResults, err := lookup.dnsRecordsResolver(
		lookupContext, lookup.secondaryResolver, hostname, dnsRecordType,
	)
	if err == nil && len(secondaryResults) > 0 {
		return secondaryResults, nil
	}

	return secondaryResults, err
}
