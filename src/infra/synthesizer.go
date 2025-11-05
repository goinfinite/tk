package tkInfra

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptoRand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	mathRand "math/rand"
	"strings"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	CharsetLowercaseLetters string = "abcdefghijklmnopqrstuvwxyz"
	CharsetUppercaseLetters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetNumbers          string = "0123456789"
	CharsetSymbols          string = "!@#$%^&*()_+"
)

type Synthesizer struct {
}

func (synth *Synthesizer) CharsetPresenceGuarantor(
	originalString []byte,
	charset string,
) []byte {
	if strings.ContainsAny(string(originalString), charset) {
		return originalString
	}

	randomStringIndex := mathRand.Intn(len(originalString))
	isFirstChar := randomStringIndex == 0
	if isFirstChar {
		randomStringIndex++
	}
	isLastChar := randomStringIndex == len(originalString)-1
	if isLastChar {
		randomStringIndex--
	}
	if randomStringIndex >= len(originalString) {
		randomStringIndex = len(originalString) - 1
	}

	randomCharsetIndex := mathRand.Intn(len(charset))
	originalString[randomStringIndex] = charset[randomCharsetIndex]

	return originalString
}

func (synth *Synthesizer) PasswordFactory(
	desiredLength int,
	shouldIncludeSymbols bool,
) string {
	alphanumericCharset := CharsetLowercaseLetters + CharsetUppercaseLetters + CharsetNumbers
	alphanumericCharsetLength := len(alphanumericCharset)

	passwordBytes := make([]byte, desiredLength)
	for charIdx := 0; charIdx < desiredLength; charIdx++ {
		passwordBytes[charIdx] = alphanumericCharset[mathRand.Intn(alphanumericCharsetLength)]
	}

	if desiredLength > 4 {
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetLowercaseLetters)
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetUppercaseLetters)
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetNumbers)
	}

	if shouldIncludeSymbols {
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetSymbols)
	}

	return string(passwordBytes)
}

func (synth *Synthesizer) UsernameFactory() string {
	dummyUsernames := []string{
		"pike", "spock", "kirk", "scotty", "bones", "uhura", "sulu", "chekov",
	}
	return dummyUsernames[mathRand.Intn(len(dummyUsernames))]
}

func (synth *Synthesizer) MailAddressFactory(username *string) string {
	if username == nil {
		dummyUsername := synth.UsernameFactory()
		username = &dummyUsername
	}

	atDomains := []string{
		"@ufp.gov", "@starfleet.gov", "@academy.edu", "@terran.gov",
	}
	return *username + atDomains[mathRand.Intn(len(atDomains))]
}

func (synth *Synthesizer) SelfSignedCertificatePairFactory(
	commonName *tkValueObject.Fqdn,
	altNames []tkValueObject.Fqdn,
) (certPair tls.Certificate, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptoRand.Reader)
	if err != nil {
		return certPair, err
	}

	validFromTime := time.Now()
	validUntilTime := validFromTime.Add(365 * 24 * time.Hour)

	certSerialNumber, err := cryptoRand.Int(
		cryptoRand.Reader, new(big.Int).Lsh(big.NewInt(1), 128),
	)
	if err != nil {
		return certPair, err
	}

	commonNameStr := "localhost"
	if commonName != nil {
		commonNameStr = commonName.String()
	}
	altNamesStrSlice := []string{}
	for _, altName := range altNames {
		altNamesStrSlice = append(altNamesStrSlice, altName.String())
	}

	certificateTemplate := x509.Certificate{
		SerialNumber: certSerialNumber,
		Subject: pkix.Name{
			Country:            []string{"Japan"},
			Organization:       []string{"Daystrom Institute"},
			OrganizationalUnit: []string{"Infrastructure Software Division"},
			Locality:           []string{"Naha"},
			Province:           []string{"Okinawa"},
			CommonName:         commonNameStr,
		},
		NotBefore:             validFromTime,
		NotAfter:              validUntilTime,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              altNamesStrSlice,
	}

	derEncodedCertBytes, err := x509.CreateCertificate(
		cryptoRand.Reader, &certificateTemplate, &certificateTemplate,
		&privateKey.PublicKey, privateKey,
	)
	if err != nil {
		return certPair, err
	}

	certPemBytes := pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: derEncodedCertBytes},
	)

	derEncodedPrivateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return certPair, err
	}

	privateKeyPemBytes := pem.EncodeToMemory(
		&pem.Block{Type: "EC PRIVATE KEY", Bytes: derEncodedPrivateKeyBytes},
	)

	return tls.X509KeyPair(certPemBytes, privateKeyPemBytes)
}

func (synth *Synthesizer) SelfSignedCertificatePairPemFactory(
	commonName *tkValueObject.Fqdn,
	altNames []tkValueObject.Fqdn,
) (certPem string, keyPem string, err error) {
	certPair, err := synth.SelfSignedCertificatePairFactory(commonName, altNames)
	if err != nil {
		return certPem, keyPem, err
	}

	certPemContent := ""
	for _, derEncodedCertBytes := range certPair.Certificate {
		pemEncodedCertBytes := pem.EncodeToMemory(
			&pem.Block{Type: "CERTIFICATE", Bytes: derEncodedCertBytes},
		)
		certPemContent += string(pemEncodedCertBytes)
	}

	assertedPrivateKey, assertOk := certPair.PrivateKey.(*ecdsa.PrivateKey)
	if !assertOk {
		return certPem, keyPem, errors.New("SelfSignedCertificatePairPrivateKeyInvalidFormat")
	}
	derEncodedPrivateKeyBytes, err := x509.MarshalECPrivateKey(assertedPrivateKey)
	if err != nil {
		return certPem, keyPem, err
	}
	keyPemContent := string(pem.EncodeToMemory(
		&pem.Block{Type: "EC PRIVATE KEY", Bytes: derEncodedPrivateKeyBytes},
	))

	return certPemContent, keyPemContent, nil
}
