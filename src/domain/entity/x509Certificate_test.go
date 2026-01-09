package entity

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestNewX509CertificateFromEnvelopedCertificate(t *testing.T) {
	synthesizer := tkInfra.Synthesizer{}

	t.Run("ValidCACertificate", func(t *testing.T) {
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{},
		)
		if err != nil {
			t.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			t.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			t.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.SerialNumber.String() == "" {
			t.Errorf("SerialNumberEmpty")
		}

		expectedVersion := uint8(3)
		if x509CertEntity.VersionNumber.Uint8() != expectedVersion {
			t.Errorf(
				"UnexpectedVersionNumber: got %d, expected %d",
				x509CertEntity.VersionNumber.Uint8(), expectedVersion,
			)
		}

		if x509CertEntity.SubjectDistinguishedName == nil {
			t.Errorf("SubjectDistinguishedNameIsNil")
		}

		if x509CertEntity.IssuerDistinguishedName == nil {
			t.Errorf("IssuerDistinguishedNameIsNil")
		}

		if x509CertEntity.ValidityNotBefore.Int64() <= 0 {
			t.Errorf("InvalidValidityNotBefore")
		}

		if x509CertEntity.ValidityNotAfter.Int64() <= 0 {
			t.Errorf("InvalidValidityNotAfter")
		}

		expectedPublicKeyAlgorithm := "ECDSA"
		if x509CertEntity.PublicKeyAlgorithm.String() != expectedPublicKeyAlgorithm {
			t.Errorf(
				"UnexpectedPublicKeyAlgorithm: got %s, expected %s",
				x509CertEntity.PublicKeyAlgorithm.String(), expectedPublicKeyAlgorithm,
			)
		}

		expectedPublicKeySize := uint16(256)
		if x509CertEntity.PublicKeySize.Uint16() != expectedPublicKeySize {
			t.Errorf(
				"UnexpectedPublicKeySize: got %d, expected %d",
				x509CertEntity.PublicKeySize.Uint16(), expectedPublicKeySize,
			)
		}

		if x509CertEntity.PublicKeyValue.String() == "" {
			t.Errorf("PublicKeyValueEmpty")
		}

		expectedSignatureAlgorithm := "ECDSAWithSHA256"
		actualSignatureAlgorithm := x509CertEntity.SignatureAlgorithm.String()
		if actualSignatureAlgorithm != expectedSignatureAlgorithm {
			t.Errorf(
				"UnexpectedSignatureAlgorithm: got %s, expected %s",
				actualSignatureAlgorithm, expectedSignatureAlgorithm,
			)
		}

		if x509CertEntity.SignatureValue == nil {
			t.Errorf("SignatureValueIsNil")
		}

		if x509CertEntity.FingerprintSHA256.String() == "" {
			t.Errorf("FingerprintSHA256Empty")
		}

		if x509CertEntity.FingerprintSHA1.String() == "" {
			t.Errorf("FingerprintSHA1Empty")
		}

		if x509CertEntity.BasicConstraints == nil {
			t.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			t.Errorf("CertificateShouldBeCA")
		}
	})

	t.Run("CACertificateWithoutMaxPathLenConstraint", func(t *testing.T) {
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{
				MaxPathLengthPtr:     nil,
				HasMaxPathLengthZero: false,
			},
		)
		if err != nil {
			t.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			t.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			t.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.BasicConstraints == nil {
			t.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			t.Errorf("CertificateShouldBeCA")
		}

		if x509CertEntity.BasicConstraints.MaxPathLength != nil {
			t.Errorf(
				"MaxPathLengthShouldBeNil: CA cert without path length "+
					"constraint should have nil MaxPathLength, got %d",
				*x509CertEntity.BasicConstraints.MaxPathLength,
			)
		}
	})

	t.Run("CACertificateWithMaxPathLenZero", func(t *testing.T) {
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{
				MaxPathLengthPtr:     nil,
				HasMaxPathLengthZero: true,
			},
		)
		if err != nil {
			t.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			t.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			t.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.BasicConstraints == nil {
			t.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			t.Errorf("CertificateShouldBeCA")
		}

		if x509CertEntity.BasicConstraints.MaxPathLength == nil {
			t.Fatalf(
				"MaxPathLengthShouldNotBeNil: CA cert with MaxPathLenZero=true " +
					"should have MaxPathLength=0",
			)
		}

		if *x509CertEntity.BasicConstraints.MaxPathLength != 0 {
			t.Errorf(
				"UnexpectedMaxPathLength: got %d, expected 0",
				*x509CertEntity.BasicConstraints.MaxPathLength,
			)
		}
	})

	t.Run("CACertificateWithMaxPathLenPositive", func(t *testing.T) {
		maxPathLen := 2
		caCertPem, _, err := synthesizer.CACertificatePemFactory(
			tkInfra.CertificateSettings{
				MaxPathLengthPtr:     &maxPathLen,
				HasMaxPathLengthZero: false,
			},
		)
		if err != nil {
			t.Fatalf("CACertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(caCertPem)
		if err != nil {
			t.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			t.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.BasicConstraints == nil {
			t.Fatalf("BasicConstraintsIsNil")
		}

		if !x509CertEntity.BasicConstraints.IsAuthority {
			t.Errorf("CertificateShouldBeCA")
		}

		if x509CertEntity.BasicConstraints.MaxPathLength == nil {
			t.Fatalf(
				"MaxPathLengthShouldNotBeNil: CA cert with MaxPathLen=%d "+
					"should have MaxPathLength set",
				maxPathLen,
			)
		}

		if *x509CertEntity.BasicConstraints.MaxPathLength != maxPathLen {
			t.Errorf(
				"UnexpectedMaxPathLength: got %d, expected %d",
				*x509CertEntity.BasicConstraints.MaxPathLength, maxPathLen,
			)
		}
	})

	t.Run("ValidSelfSignedCertificate", func(t *testing.T) {
		certPem, _, err := synthesizer.SelfSignedCertificatePairPemFactory(nil, nil)
		if err != nil {
			t.Fatalf("SelfSignedCertificateGenerationFailed: %s", err.Error())
		}

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(certPem)
		if err != nil {
			t.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		x509CertEntity, err := NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)
		if err != nil {
			t.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.SerialNumber.String() == "" {
			t.Errorf("SerialNumberEmpty")
		}

		if x509CertEntity.PublicKeyValue.String() == "" {
			t.Errorf("PublicKeyValueEmpty")
		}

		if x509CertEntity.FingerprintSHA256.String() == "" {
			t.Errorf("FingerprintSHA256Empty")
		}
	})

	t.Run("MalformedCertificate", func(t *testing.T) {
		malformedCertPem := "-----BEGIN CERTIFICATE-----\n" +
			"SGVsbG9Xb3JsZEhlbGxvV29ybGRIZWxsb1dvcmxkSGVsbG9Xb3JsZEhlbGxv\n" +
			"V29ybGRIZWxsb1dvcmxkSGVsbG9Xb3JsZEhlbGxvV29ybGRIZWxsb1dvcmxk\n" +
			"-----END CERTIFICATE-----"

		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(
			malformedCertPem,
		)
		if err != nil {
			t.Fatalf("EnvelopedCertificateCreationFailed: %s", err.Error())
		}

		_, err = NewX509CertificateFromEnvelopedCertificate(envelopedCert)
		if err == nil {
			t.Fatalf("MissingExpectedError: malformed certificate should fail")
		}

		expectedError := "ParseCertificateFailed"
		if err.Error() != expectedError {
			t.Errorf(
				"UnexpectedError: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})
}
