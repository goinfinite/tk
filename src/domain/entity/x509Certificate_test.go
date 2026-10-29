package entity_test

import (
	"errors"
	"strings"
	"testing"

	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestNewX509CertificateFromEnvelopedCertificate(test *testing.T) {
	synthesizer := tkInfra.Synthesizer{}

	test.Run("ValidCACertificate", func(subtest *testing.T) {
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{},
		)
		if err != nil {
			subtest.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			subtest.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := tkEntity.NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.SerialNumber.String() == "" {
			subtest.Errorf("SerialNumberEmpty")
		}

		expectedVersion := uint8(3)
		if x509CertEntity.VersionNumber.Uint8() != expectedVersion {
			subtest.Errorf(
				"UnexpectedVersionNumber: got %d, expected %d",
				x509CertEntity.VersionNumber.Uint8(), expectedVersion,
			)
		}

		if x509CertEntity.SubjectDistinguishedName == nil {
			subtest.Errorf("SubjectDistinguishedNameIsNil")
		}

		if x509CertEntity.IssuerDistinguishedName == nil {
			subtest.Errorf("IssuerDistinguishedNameIsNil")
		}

		if x509CertEntity.ValidityNotBefore.Int64() <= 0 {
			subtest.Errorf("InvalidValidityNotBefore")
		}

		if x509CertEntity.ValidityNotAfter.Int64() <= 0 {
			subtest.Errorf("InvalidValidityNotAfter")
		}

		expectedPublicKeyAlgorithm := "ECDSA"
		if x509CertEntity.PublicKeyAlgorithm.String() != expectedPublicKeyAlgorithm {
			subtest.Errorf(
				"UnexpectedPublicKeyAlgorithm: got %s, expected %s",
				x509CertEntity.PublicKeyAlgorithm.String(), expectedPublicKeyAlgorithm,
			)
		}

		expectedPublicKeySize := uint16(256)
		if x509CertEntity.PublicKeySize.Uint16() != expectedPublicKeySize {
			subtest.Errorf(
				"UnexpectedPublicKeySize: got %d, expected %d",
				x509CertEntity.PublicKeySize.Uint16(), expectedPublicKeySize,
			)
		}

		if x509CertEntity.PublicKeyValue.String() == "" {
			subtest.Errorf("PublicKeyValueEmpty")
		}

		expectedSignatureAlgorithm := "ECDSAWithSHA256"
		actualSignatureAlgorithm := x509CertEntity.SignatureAlgorithm.String()
		if actualSignatureAlgorithm != expectedSignatureAlgorithm {
			subtest.Errorf(
				"UnexpectedSignatureAlgorithm: got %s, expected %s",
				actualSignatureAlgorithm, expectedSignatureAlgorithm,
			)
		}

		if x509CertEntity.SignatureValue == nil {
			subtest.Errorf("SignatureValueIsNil")
		}

		if x509CertEntity.FingerprintSHA256.String() == "" {
			subtest.Errorf("FingerprintSHA256Empty")
		}

		if x509CertEntity.BasicConstraints == nil {
			subtest.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			subtest.Errorf("CertificateShouldBeCA")
		}
	})

	test.Run("CACertificateWithoutMaxPathLenConstraint", func(subtest *testing.T) {
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{
				MaxPathLengthPtr:     nil,
				HasMaxPathLengthZero: false,
			},
		)
		if err != nil {
			subtest.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			subtest.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := tkEntity.NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.BasicConstraints == nil {
			subtest.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			subtest.Errorf("CertificateShouldBeCA")
		}

		if x509CertEntity.BasicConstraints.MaxPathLength != nil {
			subtest.Errorf(
				"MaxPathLengthShouldBeNil: CA cert without path length "+
					"constraint should have nil MaxPathLength, got %d",
				*x509CertEntity.BasicConstraints.MaxPathLength,
			)
		}
	})

	test.Run("CACertificateWithMaxPathLenZero", func(subtest *testing.T) {
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{
				MaxPathLengthPtr:     nil,
				HasMaxPathLengthZero: true,
			},
		)
		if err != nil {
			subtest.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			subtest.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := tkEntity.NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.BasicConstraints == nil {
			subtest.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			subtest.Errorf("CertificateShouldBeCA")
		}

		if x509CertEntity.BasicConstraints.MaxPathLength == nil {
			subtest.Fatalf(
				"MaxPathLengthShouldNotBeNil: CA cert with MaxPathLenZero=true " +
					"should have MaxPathLength=0",
			)
		}

		if *x509CertEntity.BasicConstraints.MaxPathLength != 0 {
			subtest.Errorf(
				"UnexpectedMaxPathLength: got %d, expected 0",
				*x509CertEntity.BasicConstraints.MaxPathLength,
			)
		}
	})

	test.Run("CACertificateWithMaxPathLenPositive", func(subtest *testing.T) {
		maxPathLen := 2
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{
				MaxPathLengthPtr:     &maxPathLen,
				HasMaxPathLengthZero: false,
			},
		)
		if err != nil {
			subtest.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			subtest.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := tkEntity.NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.BasicConstraints == nil {
			subtest.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			subtest.Errorf("CertificateShouldBeCA")
		}

		if x509CertEntity.BasicConstraints.MaxPathLength == nil {
			subtest.Fatalf(
				"MaxPathLengthShouldNotBeNil: CA cert with MaxPathLen=%d "+
					"should have MaxPathLength set",
				maxPathLen,
			)
		}

		if *x509CertEntity.BasicConstraints.MaxPathLength != maxPathLen {
			subtest.Errorf(
				"UnexpectedMaxPathLength: got %d, expected %d",
				*x509CertEntity.BasicConstraints.MaxPathLength, maxPathLen,
			)
		}
	})

	test.Run("ValidSelfSignedCertificate", func(subtest *testing.T) {
		certPem, _, err := synthesizer.SelfSignedCertificatePairPemFactory(nil, nil)
		if err != nil {
			subtest.Fatalf("SelfSignedCertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(certPem)
		if err != nil {
			subtest.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := tkEntity.NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.SerialNumber.String() == "" {
			subtest.Errorf("SerialNumberEmpty")
		}

		if x509CertEntity.PublicKeyValue.String() == "" {
			subtest.Errorf("PublicKeyValueEmpty")
		}

		if x509CertEntity.FingerprintSHA256.String() == "" {
			subtest.Errorf("FingerprintSHA256Empty")
		}
	})

	test.Run("MalformedCertificate", func(subtest *testing.T) {
		malformedCertPem := "-----BEGIN CERTIFICATE-----\n" +
			"SGVsbG9Xb3JsZEhlbGxvV29ybGRIZWxsb1dvcmxkSGVsbG9Xb3JsZEhlbGxv\n" +
			"V29ybGRIZWxsb1dvcmxkSGVsbG9Xb3JsZEhlbGxvV29ybGRIZWxsb1dvcmxk\n" +
			"-----END CERTIFICATE-----"

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(
			malformedCertPem,
		)
		if err != nil {
			subtest.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		_, err = tkEntity.NewX509CertificateFromEnvelopedCertificate(envelopedCert)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: malformed certificate should fail")
		}

		expectedError := "ParseCertificateFailed"
		if !strings.Contains(err.Error(), expectedError) {
			subtest.Errorf(
				"UnexpectedError: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
		if errors.Unwrap(err) == nil {
			subtest.Errorf("ExpectedWrappedParseCertificateError")
		}
	})
}
