package tkInfra

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
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

type Synthesizer struct{}

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

type PrivateKeySettings struct {
	Algorithm tkValueObject.PrivateKeyAlgorithm
	BitSize   int
}

func (synth *Synthesizer) privateKeyGenerator(
	settings PrivateKeySettings,
) (generatedKey any, err error) {
	if settings.Algorithm == "" {
		settings.Algorithm = tkValueObject.PrivateKeyAlgorithmECDSA
	}

	switch settings.Algorithm {
	case tkValueObject.PrivateKeyAlgorithmRSA:
		bitSize := settings.BitSize
		if bitSize == 0 {
			bitSize = 2048
		}
		return rsa.GenerateKey(cryptoRand.Reader, bitSize)

	case tkValueObject.PrivateKeyAlgorithmECDSA:
		var ellipticCurve elliptic.Curve
		switch settings.BitSize {
		case 0, 256:
			ellipticCurve = elliptic.P256()
		case 384:
			ellipticCurve = elliptic.P384()
		case 521:
			ellipticCurve = elliptic.P521()
		default:
			return generatedKey, errors.New("InvalidECDSABitSize")
		}
		return ecdsa.GenerateKey(ellipticCurve, cryptoRand.Reader)

	case tkValueObject.PrivateKeyAlgorithmDSA:
		dsaParameters := new(dsa.Parameters)
		generateErr := dsa.GenerateParameters(dsaParameters, cryptoRand.Reader, dsa.L2048N256)
		if generateErr != nil {
			return generatedKey, generateErr
		}
		dsaKey := new(dsa.PrivateKey)
		dsaKey.PublicKey.Parameters = *dsaParameters
		generateErr = dsa.GenerateKey(dsaKey, cryptoRand.Reader)
		if generateErr != nil {
			return generatedKey, generateErr
		}
		return dsaKey, nil

	case tkValueObject.PrivateKeyAlgorithmEd25519:
		_, ed25519Key, generateErr := ed25519.GenerateKey(cryptoRand.Reader)
		if generateErr != nil {
			return generatedKey, generateErr
		}
		return ed25519Key, nil

	default:
		return generatedKey, errors.New("UnsupportedPrivateKeyAlgorithm")
	}
}

func (synth *Synthesizer) PrivateKeyPemFactory(
	settings PrivateKeySettings,
) (keyPem string, err error) {
	generatedPrivateKey, err := synth.privateKeyGenerator(settings)
	if err != nil {
		return keyPem, err
	}

	var pemBlockType string
	var derEncodedBytes []byte

	switch generatedPrivateKey := generatedPrivateKey.(type) {
	case *rsa.PrivateKey:
		derEncodedBytes = x509.MarshalPKCS1PrivateKey(generatedPrivateKey)
		pemBlockType = "RSA PRIVATE KEY"

	case *ecdsa.PrivateKey:
		derEncodedBytes, err = x509.MarshalECPrivateKey(generatedPrivateKey)
		if err != nil {
			return keyPem, err
		}
		pemBlockType = "EC PRIVATE KEY"

	case *dsa.PrivateKey:
		type dsaPrivateKeyAsn1 struct {
			Version       int
			P, Q, G, Y, X *big.Int
		}

		asn1Representation := dsaPrivateKeyAsn1{
			Version: 0,
			P:       generatedPrivateKey.P,
			Q:       generatedPrivateKey.Q,
			G:       generatedPrivateKey.G,
			Y:       generatedPrivateKey.Y,
			X:       generatedPrivateKey.X,
		}
		derEncodedBytes, err = asn1.Marshal(asn1Representation)
		if err != nil {
			return keyPem, err
		}
		pemBlockType = "DSA PRIVATE KEY"

	case ed25519.PrivateKey:
		derEncodedBytes, err = x509.MarshalPKCS8PrivateKey(generatedPrivateKey)
		if err != nil {
			return keyPem, err
		}
		pemBlockType = "PRIVATE KEY"

	default:
		return keyPem, errors.New("UnsupportedPrivateKeyType")
	}

	pemEncodedBytes := pem.EncodeToMemory(
		&pem.Block{Type: pemBlockType, Bytes: derEncodedBytes},
	)

	return string(pemEncodedBytes), nil
}

type CertificateSettings struct {
	CommonName           *tkValueObject.Fqdn
	AltNames             []tkValueObject.Fqdn
	IsCA                 bool
	MaxPathLengthPtr     *int
	HasMaxPathLengthZero bool
}

func (synth *Synthesizer) certTemplateGenerator(
	settings CertificateSettings,
) (template x509.Certificate, serialNumber *big.Int, err error) {
	// The serial number is a positive integer that must be unique for every certificate issued
	// by a given CA. This method generates a random non-negative integer in the range [0, 2^128).
	serialNumber, err = cryptoRand.Int(
		cryptoRand.Reader,
		new(big.Int).Lsh(big.NewInt(1), 128),
	)
	if err != nil {
		return template, serialNumber, err
	}

	commonNameValue := "localhost"
	if settings.IsCA {
		commonNameValue = "Test CA"
	}
	if settings.CommonName != nil {
		commonNameValue = settings.CommonName.String()
	}

	subjectAlternativeNames := []string{}
	for _, altName := range settings.AltNames {
		subjectAlternativeNames = append(subjectAlternativeNames, altName.String())
	}

	organizationalUnit := "Infrastructure Software Division"
	if settings.IsCA {
		organizationalUnit = "Certificate Authority"
	}

	validFromTime := time.Now()
	validUntilTime := validFromTime.Add(365 * 24 * time.Hour)

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if settings.IsCA {
		keyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	maxPathLen := 0
	if settings.MaxPathLengthPtr != nil {
		maxPathLen = *settings.MaxPathLengthPtr
	}

	template = x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"JP"},
			Organization:       []string{"Daystrom Institute"},
			OrganizationalUnit: []string{organizationalUnit},
			Locality:           []string{"Naha"},
			Province:           []string{"Okinawa"},
			CommonName:         commonNameValue,
		},
		NotBefore:             validFromTime,
		NotAfter:              validUntilTime,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  settings.IsCA,
		MaxPathLen:            maxPathLen,
		MaxPathLenZero:        settings.HasMaxPathLengthZero,
		DNSNames:              subjectAlternativeNames,
	}

	return template, serialNumber, nil
}

func (synth *Synthesizer) selfSignedCertPemBytesGenerator(
	settings CertificateSettings,
) (certPemBytes []byte, keyPemBytes []byte, err error) {
	generatedPrivateKey, err := synth.privateKeyGenerator(PrivateKeySettings{
		Algorithm: tkValueObject.PrivateKeyAlgorithmECDSA,
		BitSize:   256,
	})
	if err != nil {
		return certPemBytes, keyPemBytes, err
	}

	ecdsaPrivateKey, assertOk := generatedPrivateKey.(*ecdsa.PrivateKey)
	if !assertOk {
		return certPemBytes, keyPemBytes, errors.New("PrivateKeyAssertionFailed")
	}

	certificateTemplate, _, err := synth.certTemplateGenerator(settings)
	if err != nil {
		return certPemBytes, keyPemBytes, err
	}

	derEncodedCertBytes, err := x509.CreateCertificate(
		cryptoRand.Reader, &certificateTemplate, &certificateTemplate,
		&ecdsaPrivateKey.PublicKey, ecdsaPrivateKey,
	)
	if err != nil {
		return certPemBytes, keyPemBytes, err
	}

	certPemBytes = pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: derEncodedCertBytes},
	)

	derEncodedPrivateKeyBytes, err := x509.MarshalECPrivateKey(ecdsaPrivateKey)
	if err != nil {
		return certPemBytes, keyPemBytes, err
	}

	keyPemBytes = pem.EncodeToMemory(
		&pem.Block{Type: "EC PRIVATE KEY", Bytes: derEncodedPrivateKeyBytes},
	)

	return certPemBytes, keyPemBytes, nil
}

func (synth *Synthesizer) CertificatePemFactory(
	settings CertificateSettings,
) (certPem string, keyPem string, err error) {
	certPemBytes, keyPemBytes, err := synth.selfSignedCertPemBytesGenerator(settings)
	if err != nil {
		return certPem, keyPem, err
	}

	return string(certPemBytes), string(keyPemBytes), nil
}

func (synth *Synthesizer) CACertificatePemFactory(
	settings CertificateSettings,
) (certPem string, keyPem string, err error) {
	settings.IsCA = true
	return synth.CertificatePemFactory(settings)
}

func (synth *Synthesizer) SelfSignedCertificatePairFactory(
	commonName *tkValueObject.Fqdn,
	altNames []tkValueObject.Fqdn,
) (certPair tls.Certificate, err error) {
	certPemBytes, keyPemBytes, err := synth.selfSignedCertPemBytesGenerator(
		CertificateSettings{
			CommonName: commonName,
			AltNames:   altNames,
			IsCA:       false,
		},
	)
	if err != nil {
		return certPair, err
	}

	return tls.X509KeyPair(certPemBytes, keyPemBytes)
}

func (synth *Synthesizer) SelfSignedCertificatePairPemFactory(
	commonName *tkValueObject.Fqdn,
	altNames []tkValueObject.Fqdn,
) (certPem string, keyPem string, err error) {
	certPair, err := synth.SelfSignedCertificatePairFactory(commonName, altNames)
	if err != nil {
		return certPem, keyPem, err
	}

	var certPemContent strings.Builder
	for _, derEncodedCertBytes := range certPair.Certificate {
		pemEncodedCertBytes := pem.EncodeToMemory(
			&pem.Block{Type: "CERTIFICATE", Bytes: derEncodedCertBytes},
		)
		certPemContent.WriteString(string(pemEncodedCertBytes))
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

	return certPemContent.String(), keyPemContent, nil
}
