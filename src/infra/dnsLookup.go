package tkInfra

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log/slog"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	dnsLookupQueryTimeoutSecsDefault uint = 5
	dnsLookupDialTimeoutMsDefault    uint = 200

	dnsStandardPort = "53"
)

var (
	dnsLookupPrimaryResolverDefault   = tkValueObject.IpAddress("8.8.8.8")
	dnsLookupSecondaryResolverDefault = tkValueObject.IpAddress("185.228.168.168")

	ErrDnsLookupResponseIdMismatch    = errors.New("DnsLookupResponseIdMismatch")
	ErrDnsLookupResponseNameError     = errors.New("DnsLookupResponseNameError")
	ErrDnsLookupResponseServerFailure = errors.New("DnsLookupResponseServerFailure")
	ErrDnsLookupResponseRefused       = errors.New("DnsLookupResponseRefused")
	ErrDnsLookupResponseUnknownRCode  = errors.New("DnsLookupResponseUnknownRCode")
	ErrDnsLookupResponseTruncated     = errors.New("DnsLookupResponseTruncated")
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

func (lookup *DnsLookup) netResolverBuilder(
	resolverIpAddress tkValueObject.IpAddress,
) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{
				Timeout: time.Duration(lookup.dialTimeoutMs) * time.Millisecond,
			}
			return dialer.DialContext(
				ctx, "udp", resolverIpAddress.String()+":"+dnsStandardPort,
			)
		},
	}
}

func (lookup *DnsLookup) dnsMessagePacker(
	hostname tkValueObject.UnixHostname,
	questionType dnsmessage.Type,
) (queryBytes []byte, transactionId uint16, buildError error) {
	dnsName, nameParseError := dnsmessage.NewName(hostname.String() + ".")
	if nameParseError != nil {
		return nil, 0, nameParseError
	}

	var idBytes [2]byte
	_, readError := rand.Read(idBytes[:])
	if readError != nil {
		return nil, 0, errors.New(
			"DnsLookupQueryIdRandomSourceUnavailable: " + readError.Error(),
		)
	}
	transactionId = binary.BigEndian.Uint16(idBytes[:])

	queryMessage := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:               transactionId,
			RecursionDesired: true,
		},
		Questions: []dnsmessage.Question{{
			Name:  dnsName,
			Type:  questionType,
			Class: dnsmessage.ClassINET,
		}},
	}

	packedBytes, packError := queryMessage.Pack()
	if packError != nil {
		return nil, 0, packError
	}

	return packedBytes, transactionId, nil
}

func (lookup *DnsLookup) dnsMessageExchanger(
	dnsContext context.Context,
	resolverIpAddress tkValueObject.IpAddress,
	queryBytes []byte,
) (responseBytes []byte, exchangeError error) {
	dialer := net.Dialer{
		Timeout: time.Duration(lookup.dialTimeoutMs) * time.Millisecond,
	}
	udpConn, dialError := dialer.DialContext(
		dnsContext, "udp", resolverIpAddress.String()+":"+dnsStandardPort,
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

func (lookup *DnsLookup) dnsMessageValidator(
	responseBytes []byte,
	expectedTransactionId uint16,
) (responseMessage dnsmessage.Message, err error) {
	unpackFailure := responseMessage.Unpack(responseBytes)
	if unpackFailure != nil {
		err = unpackFailure
		return
	}

	if responseMessage.Header.ID != expectedTransactionId {
		err = ErrDnsLookupResponseIdMismatch
		return
	}

	switch responseMessage.Header.RCode {
	case dnsmessage.RCodeSuccess:
	case dnsmessage.RCodeNameError:
		err = ErrDnsLookupResponseNameError
		return
	case dnsmessage.RCodeServerFailure:
		err = ErrDnsLookupResponseServerFailure
		return
	case dnsmessage.RCodeRefused:
		err = ErrDnsLookupResponseRefused
		return
	default:
		err = ErrDnsLookupResponseUnknownRCode
		return
	}

	if responseMessage.Header.Truncated {
		err = ErrDnsLookupResponseTruncated
		return
	}

	return
}

func (lookup *DnsLookup) dnsMessageIpAddrExtractor(
	responseMessage dnsmessage.Message,
) (ipAddresses []string) {
	for _, answer := range responseMessage.Answers {
		switch typedAnswer := answer.Body.(type) {
		case *dnsmessage.AResource:
			ipAddresses = append(
				ipAddresses, net.IP(typedAnswer.A[:]).String(),
			)
		case *dnsmessage.AAAAResource:
			ipAddresses = append(
				ipAddresses, net.IP(typedAnswer.AAAA[:]).String(),
			)
		}
	}
	return
}

func (lookup *DnsLookup) directIpAddressResolver(
	dnsContext context.Context,
	resolverIpAddress tkValueObject.IpAddress,
	hostname tkValueObject.UnixHostname,
	recordType tkValueObject.DnsRecordType,
) ([]string, error) {
	var questionType dnsmessage.Type
	switch recordType {
	case tkValueObject.DnsRecordTypeA:
		questionType = dnsmessage.TypeA
	case tkValueObject.DnsRecordTypeAAAA:
		questionType = dnsmessage.TypeAAAA
	default:
		return nil, errors.New(
			"DnsLookupDirectResolutionUnsupportedRecordType: " + recordType.String(),
		)
	}

	queryBytes, transactionId, buildError := lookup.dnsMessagePacker(
		hostname, questionType,
	)
	if buildError != nil {
		return nil, buildError
	}

	responseBytes, exchangeError := lookup.dnsMessageExchanger(
		dnsContext, resolverIpAddress, queryBytes,
	)
	if exchangeError != nil {
		return nil, exchangeError
	}

	responseMessage, validateError := lookup.dnsMessageValidator(
		responseBytes, transactionId,
	)
	if validateError != nil {
		return nil, validateError
	}

	return lookup.dnsMessageIpAddrExtractor(responseMessage), nil
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
			parsedIp := net.ParseIP(dnsRecord)
			if parsedIp != nil && parsedIp.To4() == nil {
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
	case tkValueObject.DnsRecordTypeCNAME:
		cnameRecord, err := dnsResolver.LookupCNAME(dnsContext, hostnameStr)
		if err != nil {
			return nil, err
		}
		queryResults = []string{cnameRecord}
	case tkValueObject.DnsRecordTypePTR:
		ptrRecords, err := dnsResolver.LookupAddr(dnsContext, hostnameStr)
		if err != nil {
			return nil, err
		}
		queryResults = ptrRecords
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

	resolver := lookup.netResolverBuilder(resolverIpAddress)
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

	resolverIpAddresses := []tkValueObject.IpAddress{
		lookup.primaryResolver, lookup.secondaryResolver,
	}

	var lastRecords []string
	var lastError error
	for _, resolverIpAddress := range resolverIpAddresses {
		attemptContext, attemptCancel := context.WithTimeout(
			context.Background(),
			time.Duration(lookup.queryTimeoutSecs)*time.Second,
		)
		records, lookupError := lookup.dnsRecordsResolver(
			attemptContext, resolverIpAddress, hostname, dnsRecordType,
		)
		attemptCancel()

		if lookupError == nil && len(records) > 0 {
			return records, nil
		}

		if lookupError != nil {
			slog.Debug(
				"DnsLookupResolverFailed",
				slog.String("hostname", hostname.String()),
				slog.String("resolverIpAddress", resolverIpAddress.String()),
				slog.String("error", lookupError.Error()),
			)
		}

		lastRecords = records
		lastError = lookupError
	}

	return lastRecords, lastError
}
