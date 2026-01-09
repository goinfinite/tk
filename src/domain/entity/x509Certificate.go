package entity

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"log/slog"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type X509Certificate struct {
	SerialNumber  tkValueObject.X509SerialNumber  `json:"serialNumber"`
	VersionNumber tkValueObject.X509VersionNumber `json:"versionNumber"`

	SubjectCommonName        *tkValueObject.X509SubjectName       `json:"subjectCommonName"`
	SubjectAltNames          []tkValueObject.X509SubjectName      `json:"subjectAltNames"`
	SubjectDistinguishedName *tkValueObject.X509DistinguishedName `json:"subjectDistinguishedName"`

	IssuerCommonName        *tkValueObject.X509SubjectName       `json:"issuerCommonName"`
	IssuerDistinguishedName *tkValueObject.X509DistinguishedName `json:"issuerDistinguishedName"`

	ValidityNotBefore tkValueObject.UnixTime `json:"validityNotBefore"`
	ValidityNotAfter  tkValueObject.UnixTime `json:"validityNotAfter"`

	PublicKeyAlgorithm tkValueObject.X509PublicKeyAlgorithm `json:"publicKeyAlgorithm"`
	PublicKeySize      tkValueObject.X509PublicKeySize      `json:"publicKeySize"`
	PublicKeyValue     tkValueObject.X509PublicKeyValue     `json:"publicKeyValue"`

	SignatureAlgorithm tkValueObject.X509SignatureAlgorithm `json:"signatureAlgorithm"`
	SignatureValue     *tkValueObject.X509SignatureValue    `json:"signatureValue"`

	FingerprintSHA256 tkValueObject.X509Fingerprint `json:"fingerprintSha256"`
	FingerprintSHA1   tkValueObject.X509Fingerprint `json:"fingerprintSha1"`

	KeyUsage               []tkValueObject.X509KeyUsage         `json:"keyUsage"`
	ExtendedKeyUsage       []tkValueObject.X509ExtendedKeyUsage `json:"extendedKeyUsage"`
	BasicConstraints       *tkValueObject.X509BasicConstraints  `json:"basicConstraints"`
	SubjectKeyIdentifier   *tkValueObject.X509KeyIdentifier     `json:"subjectKeyIdentifier"`
	AuthorityKeyIdentifier *tkValueObject.X509KeyIdentifier     `json:"authorityKeyIdentifier"`

	CertificatePolicies []tkValueObject.X509CertificatePolicy `json:"certificatePolicies"`

	EnvelopedCertificate tkValueObject.X509EnvelopedCertificate `json:"envelopedCertificate"`
}

func NewX509Certificate(
	serialNumber tkValueObject.X509SerialNumber,
	versionNumber tkValueObject.X509VersionNumber,
	subjectCommonName *tkValueObject.X509SubjectName,
	subjectAltNames []tkValueObject.X509SubjectName,
	subjectDistinguishedName *tkValueObject.X509DistinguishedName,
	issuerCommonName *tkValueObject.X509SubjectName,
	issuerDistinguishedName *tkValueObject.X509DistinguishedName,
	validityNotBefore, validityNotAfter tkValueObject.UnixTime,
	publicKeyAlgorithm tkValueObject.X509PublicKeyAlgorithm,
	publicKeySize tkValueObject.X509PublicKeySize,
	publicKeyValue tkValueObject.X509PublicKeyValue,
	signatureAlgorithm tkValueObject.X509SignatureAlgorithm,
	signatureValue *tkValueObject.X509SignatureValue,
	fingerprintSHA256, fingerprintSHA1 tkValueObject.X509Fingerprint,
	keyUsage []tkValueObject.X509KeyUsage,
	extendedKeyUsage []tkValueObject.X509ExtendedKeyUsage,
	basicConstraints *tkValueObject.X509BasicConstraints,
	subjectKeyIdentifier, authorityKeyIdentifier *tkValueObject.X509KeyIdentifier,
	certificatePolicies []tkValueObject.X509CertificatePolicy,
	envelopedCertificate tkValueObject.X509EnvelopedCertificate,
) X509Certificate {
	return X509Certificate{
		SerialNumber:             serialNumber,
		VersionNumber:            versionNumber,
		SubjectCommonName:        subjectCommonName,
		SubjectAltNames:          subjectAltNames,
		SubjectDistinguishedName: subjectDistinguishedName,
		IssuerCommonName:         issuerCommonName,
		IssuerDistinguishedName:  issuerDistinguishedName,
		ValidityNotBefore:        validityNotBefore,
		ValidityNotAfter:         validityNotAfter,
		PublicKeyAlgorithm:       publicKeyAlgorithm,
		PublicKeySize:            publicKeySize,
		PublicKeyValue:           publicKeyValue,
		SignatureAlgorithm:       signatureAlgorithm,
		SignatureValue:           signatureValue,
		FingerprintSHA256:        fingerprintSHA256,
		FingerprintSHA1:          fingerprintSHA1,
		KeyUsage:                 keyUsage,
		ExtendedKeyUsage:         extendedKeyUsage,
		BasicConstraints:         basicConstraints,
		SubjectKeyIdentifier:     subjectKeyIdentifier,
		AuthorityKeyIdentifier:   authorityKeyIdentifier,
		CertificatePolicies:      certificatePolicies,
		EnvelopedCertificate:     envelopedCertificate,
	}
}

func NewX509CertificateFromEnvelopedCertificate(
	envelopedCertificate tkValueObject.X509EnvelopedCertificate,
) (x509CertEntity X509Certificate, err error) {
	envelopedCertificateBytes := envelopedCertificate.Bytes()
	pemBlock, remainingBytes := pem.Decode(envelopedCertificateBytes)
	if pemBlock == nil {
		return x509CertEntity, errors.New("DecodePEMFailed")
	}

	if len(remainingBytes) > 0 {
		return x509CertEntity, errors.New("MultipleX509CertificatesNotAllowed")
	}

	stdlibCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return x509CertEntity, errors.New("ParseCertificateFailed")
	}

	serialNumberHex := stdlibCert.SerialNumber.Text(16)
	serialNumber, err := tkValueObject.NewX509SerialNumber(serialNumberHex)
	if err != nil {
		return x509CertEntity, err
	}

	certVersion := uint8(stdlibCert.Version)
	versionNumber, err := tkValueObject.NewX509VersionNumber(certVersion)
	if err != nil {
		return x509CertEntity, err
	}

	var subjectCommonNamePtr *tkValueObject.X509SubjectName
	rawSubjectCommonName := stdlibCert.Subject.CommonName
	if rawSubjectCommonName != "" {
		subjectCommonName, err := tkValueObject.NewX509SubjectName(
			rawSubjectCommonName,
		)
		if err != nil {
			return x509CertEntity, err
		}
		subjectCommonNamePtr = &subjectCommonName
	}

	var subjectAltNames []tkValueObject.X509SubjectName
	for _, dnsName := range stdlibCert.DNSNames {
		subjectAltName, err := tkValueObject.NewX509SubjectName(dnsName)
		if err != nil {
			slog.Debug(
				"SkipInvalidSubjectAltName",
				slog.String("dnsName", dnsName),
			)
			continue
		}
		subjectAltNames = append(subjectAltNames, subjectAltName)
	}

	subjectDistinguishedNamePtr, err :=
		tkValueObject.NewX509DistinguishedNameFromPkixName(stdlibCert.Subject)
	if err != nil {
		return x509CertEntity, err
	}

	var issuerCommonNamePtr *tkValueObject.X509SubjectName
	rawIssuerCommonName := stdlibCert.Issuer.CommonName
	if rawIssuerCommonName != "" {
		issuerCommonName, err := tkValueObject.NewX509SubjectName(
			rawIssuerCommonName,
		)
		if err != nil {
			return x509CertEntity, err
		}
		issuerCommonNamePtr = &issuerCommonName
	}

	issuerDistinguishedNamePtr, err :=
		tkValueObject.NewX509DistinguishedNameFromPkixName(stdlibCert.Issuer)
	if err != nil {
		return x509CertEntity, err
	}

	validityNotBeforeUnix := stdlibCert.NotBefore.Unix()
	validityNotBefore, err := tkValueObject.NewUnixTime(validityNotBeforeUnix)
	if err != nil {
		return x509CertEntity, err
	}

	validityNotAfterUnix := stdlibCert.NotAfter.Unix()
	validityNotAfter, err := tkValueObject.NewUnixTime(validityNotAfterUnix)
	if err != nil {
		return x509CertEntity, err
	}

	publicKeyAlgorithm, err :=
		tkValueObject.NewX509PublicKeyAlgorithmFromStdlib(
			stdlibCert.PublicKeyAlgorithm,
		)
	if err != nil {
		return x509CertEntity, err
	}

	publicKeySize, err := tkValueObject.NewX509PublicKeySizeFromStdlib(
		stdlibCert.PublicKey,
	)
	if err != nil {
		return x509CertEntity, err
	}

	publicKeyHex := hex.EncodeToString(stdlibCert.RawSubjectPublicKeyInfo)
	publicKeyValue, err := tkValueObject.NewX509PublicKeyValue(publicKeyHex)
	if err != nil {
		return x509CertEntity, err
	}

	signatureAlgorithm, err :=
		tkValueObject.NewX509SignatureAlgorithmFromStdlib(
			stdlibCert.SignatureAlgorithm,
		)
	if err != nil {
		return x509CertEntity, err
	}

	signatureHex := hex.EncodeToString(stdlibCert.Signature)
	signatureValue, err := tkValueObject.NewX509SignatureValue(signatureHex)
	if err != nil {
		return x509CertEntity, err
	}
	signatureValuePtr := &signatureValue

	sha256HashBytes := sha256.Sum256(stdlibCert.Raw)
	sha256FingerprintHex := hex.EncodeToString(sha256HashBytes[:])
	fingerprintSHA256, err := tkValueObject.NewX509Fingerprint(
		sha256FingerprintHex,
	)
	if err != nil {
		return x509CertEntity, err
	}

	sha1HashBytes := sha1.Sum(stdlibCert.Raw)
	sha1FingerprintHex := hex.EncodeToString(sha1HashBytes[:])
	fingerprintSHA1, err := tkValueObject.NewX509Fingerprint(sha1FingerprintHex)
	if err != nil {
		return x509CertEntity, err
	}

	keyUsageSlice, err := tkValueObject.NewX509KeyUsageSliceFromStdlib(
		stdlibCert.KeyUsage,
	)
	if err != nil {
		return x509CertEntity, err
	}

	extendedKeyUsageSlice, err :=
		tkValueObject.NewX509ExtendedKeyUsageSliceFromStdlib(
			stdlibCert.ExtKeyUsage,
		)
	if err != nil {
		return x509CertEntity, err
	}

	var basicConstraintsPtr *tkValueObject.X509BasicConstraints
	basicConstraintsAreValid := stdlibCert.BasicConstraintsValid
	if basicConstraintsAreValid {
		var maxPathLengthPtr *int
		certIsAuthority := stdlibCert.IsCA
		caHasMaxPathLengthConstraint := stdlibCert.MaxPathLen > 0 ||
			(stdlibCert.MaxPathLen == 0 && stdlibCert.MaxPathLenZero)
		if certIsAuthority && caHasMaxPathLengthConstraint {
			maxPathLengthPtr = &stdlibCert.MaxPathLen
		}
		basicConstraints, err := tkValueObject.NewX509BasicConstraints(
			certIsAuthority, maxPathLengthPtr,
		)
		if err != nil {
			return x509CertEntity, err
		}
		basicConstraintsPtr = &basicConstraints
	}

	var subjectKeyIdentifierPtr *tkValueObject.X509KeyIdentifier
	if len(stdlibCert.SubjectKeyId) > 0 {
		subjectKeyIdHex := hex.EncodeToString(stdlibCert.SubjectKeyId)
		subjectKeyIdentifier, err := tkValueObject.NewX509KeyIdentifier(
			subjectKeyIdHex,
		)
		if err == nil {
			subjectKeyIdentifierPtr = &subjectKeyIdentifier
		}
	}

	var authorityKeyIdentifierPtr *tkValueObject.X509KeyIdentifier
	if len(stdlibCert.AuthorityKeyId) > 0 {
		authorityKeyIdHex := hex.EncodeToString(stdlibCert.AuthorityKeyId)
		authorityKeyIdentifier, err := tkValueObject.NewX509KeyIdentifier(
			authorityKeyIdHex,
		)
		if err == nil {
			authorityKeyIdentifierPtr = &authorityKeyIdentifier
		}
	}

	var certificatePolicies []tkValueObject.X509CertificatePolicy
	for _, stdlibPolicyOID := range stdlibCert.PolicyIdentifiers {
		policyOID, err := tkValueObject.NewX509PolicyOID(stdlibPolicyOID.String())
		if err != nil {
			slog.Debug(
				"SkipInvalidPolicyOID",
				slog.String("oid", stdlibPolicyOID.String()),
			)
			continue
		}
		certificatePolicy := tkValueObject.NewX509CertificatePolicy(
			policyOID, nil, nil,
		)
		certificatePolicies = append(certificatePolicies, certificatePolicy)
	}

	return NewX509Certificate(
		serialNumber, versionNumber, subjectCommonNamePtr, subjectAltNames,
		subjectDistinguishedNamePtr, issuerCommonNamePtr,
		issuerDistinguishedNamePtr, validityNotBefore, validityNotAfter,
		publicKeyAlgorithm, publicKeySize, publicKeyValue, signatureAlgorithm,
		signatureValuePtr, fingerprintSHA256, fingerprintSHA1, keyUsageSlice,
		extendedKeyUsageSlice, basicConstraintsPtr, subjectKeyIdentifierPtr,
		authorityKeyIdentifierPtr, certificatePolicies, envelopedCertificate,
	), nil
}
