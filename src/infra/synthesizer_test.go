package tkInfra

import (
	"crypto/x509"
	"encoding/pem"
	"regexp"
	"strings"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestCharsetPresenceGuarantor(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("CharsetAlreadyPresent", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    []byte
			charset        string
			expectedOutput []byte
		}{
			{[]byte("abc123"), "123", []byte("abc123")},
			{[]byte("ABCDEF"), "ABCDEF", []byte("ABCDEF")},
			{[]byte("abc!@#"), "!@#", []byte("abc!@#")},
			{[]byte("a"), "a", []byte("a")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := synth.CharsetPresenceGuarantor(testCase.inputString, testCase.charset)
			if string(actualOutput) != string(testCase.expectedOutput) {
				t.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' [%s, %s]",
					string(actualOutput), string(testCase.expectedOutput), string(testCase.inputString),
					testCase.charset,
				)
			}
		}
	})

	t.Run("CharsetNotPresent", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString []byte
			charset     string
		}{
			{[]byte("abcdef"), "123"},
			{[]byte("123456"), "abc"},
			{[]byte("abcABC"), "!@#"},
		}

		for _, testCase := range testCaseStructs {
			inputCopy := make([]byte, len(testCase.inputString))
			copy(inputCopy, testCase.inputString)

			actualOutput := synth.CharsetPresenceGuarantor(inputCopy, testCase.charset)

			if string(actualOutput) == string(testCase.inputString) {
				t.Errorf(
					"OutputUnchanged: Expected modification to include charset '%s' in '%s'",
					testCase.charset, string(testCase.inputString),
				)
			}

			if !strings.ContainsAny(string(actualOutput), testCase.charset) {
				t.Errorf(
					"CharsetNotAdded: Output '%s' does not contain any character from charset '%s'",
					string(actualOutput), testCase.charset,
				)
			}

			differentChars := 0
			for i := 0; i < len(testCase.inputString); i++ {
				if actualOutput[i] != testCase.inputString[i] {
					differentChars++
				}
			}
			if differentChars != 1 {
				t.Errorf(
					"UnexpectedModificationCount: %d characters modified instead of 1 [%s, %s]",
					differentChars, string(testCase.inputString), testCase.charset,
				)
			}
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		singleChar := []byte("a")
		charset := "b"
		actualOutput := synth.CharsetPresenceGuarantor(singleChar, charset)
		if string(actualOutput) != "b" {
			t.Errorf(
				"UnexpectedOutputValue: '%s' vs 'b' [single character case]",
				string(actualOutput),
			)
		}

		twoChars := []byte("ab")
		charset = "c"
		actualOutput = synth.CharsetPresenceGuarantor(twoChars, charset)
		if !strings.ContainsAny(string(actualOutput), charset) {
			t.Errorf(
				"CharsetNotAdded: Output '%s' does not contain any character from charset '%s'",
				string(actualOutput), charset,
			)
		}
	})
}

func TestPasswordFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("PasswordLength", func(t *testing.T) {
		testCaseStructs := []struct {
			desiredLength        int
			shouldIncludeSymbols bool
		}{
			{8, false},
			{12, false},
			{16, true},
			{20, true},
			{4, false}, // Edge case: small length
			{3, true},  // Edge case: very small length
			{0, false}, // Edge case: zero length
		}

		for _, testCase := range testCaseStructs {
			password := synth.PasswordFactory(
				testCase.desiredLength, testCase.shouldIncludeSymbols,
			)
			if len(password) != testCase.desiredLength {
				t.Errorf(
					"UnexpectedPasswordLength: '%d' vs '%d' [desired: %d, includeSymbols: %t]",
					len(password), testCase.desiredLength, testCase.desiredLength,
					testCase.shouldIncludeSymbols,
				)
			}
		}
	})

	t.Run("PasswordCharacteristics", func(t *testing.T) {
		password := synth.PasswordFactory(12, false)
		if !strings.ContainsAny(password, CharsetLowercaseLetters) {
			t.Errorf("MissingLowercaseLetters: '%s'", password)
		}

		if !strings.ContainsAny(password, CharsetUppercaseLetters) {
			t.Errorf("MissingUppercaseLetters: '%s'", password)
		}

		if !strings.ContainsAny(password, CharsetNumbers) {
			t.Errorf("MissingNumbers: '%s'", password)
		}

		if len(password) != 12 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '12'", len(password))
		}

		passwordWithSymbols := synth.PasswordFactory(12, true)
		if !strings.ContainsAny(passwordWithSymbols, CharsetSymbols) {
			t.Errorf("MissingSymbols: '%s'", passwordWithSymbols)
		}

		if len(passwordWithSymbols) != 12 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '12'", len(passwordWithSymbols))
		}
	})

	t.Run("ShortPasswordCharacteristics", func(t *testing.T) {
		shortPassword := synth.PasswordFactory(4, false)
		if len(shortPassword) != 4 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '4'", len(shortPassword))
		}

		veryShortPassword := synth.PasswordFactory(3, true)
		if len(veryShortPassword) != 3 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '3'", len(veryShortPassword))
		}
	})
}

func TestUsernameFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("UsernameGeneration", func(t *testing.T) {
		usernamesRegex := `^\w{1,256}$`
		re := regexp.MustCompile(usernamesRegex)

		for range 5 {
			username := synth.UsernameFactory()
			if !re.MatchString(username) {
				t.Errorf(
					"InvalidUsernameFormat: '%s' does not match regex '%s'",
					username, usernamesRegex,
				)
			}
		}
	})
}

func TestMailAddressFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("WithNilUsername", func(t *testing.T) {
		rawMailAddress := synth.MailAddressFactory(nil)
		_, err := tkValueObject.NewMailAddress(rawMailAddress)
		if err != nil {
			t.Errorf("InvalidMailAddress: '%s' is not a valid email address", rawMailAddress)
		}
	})

	t.Run("WithProvidedUsername", func(t *testing.T) {
		testCaseStructs := []struct {
			username string
		}{
			{"testuser"},
			{"admin"},
			{"user123"},
		}

		for _, testCase := range testCaseStructs {
			username := testCase.username
			rawMailAddress := synth.MailAddressFactory(&username)
			mailAddress, err := tkValueObject.NewMailAddress(rawMailAddress)
			if err != nil {
				t.Errorf("InvalidMailAddress: '%s' is not a valid email address", rawMailAddress)
			}

			if !strings.HasPrefix(mailAddress.String(), username) {
				t.Errorf("MissingUsername: '%s' does not start with '%s'", mailAddress, username)
			}
		}
	})
}

func TestSelfSignedCertificatePairFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("WithNilCommonNameAndEmptyAltNames", func(t *testing.T) {
		certPair, err := synth.SelfSignedCertificatePairFactory(nil, []tkValueObject.Fqdn{})
		if err != nil {
			t.Errorf("UnexpectedError: %v", err)
		}

		if certPair.Leaf == nil {
			t.Error("CertificateLeafIsNil")
		}

		if certPair.Leaf.Subject.CommonName != "localhost" {
			t.Errorf("UnexpectedCommonName: '%s' vs 'localhost'", certPair.Leaf.Subject.CommonName)
		}

		if len(certPair.Leaf.DNSNames) != 0 {
			t.Errorf("UnexpectedAltNames: %v", certPair.Leaf.DNSNames)
		}

		if certPair.Leaf.Subject.Organization[0] != "Daystrom Institute" {
			t.Errorf("UnexpectedOrganization: '%s' vs 'Daystrom Institute'", certPair.Leaf.Subject.Organization[0])
		}

		if certPair.Leaf.SerialNumber == nil || certPair.Leaf.SerialNumber.Sign() <= 0 {
			t.Error("InvalidSerialNumber")
		}

		validFrom := certPair.Leaf.NotBefore
		validUntil := certPair.Leaf.NotAfter
		if validUntil.Sub(validFrom) != 365*24*time.Hour {
			t.Errorf("UnexpectedValidityPeriod: %v", validUntil.Sub(validFrom))
		}
	})

	t.Run("WithProvidedCommonName", func(t *testing.T) {
		commonName, err := tkValueObject.NewFqdn("example.com")
		if err != nil {
			t.Fatalf("CreateFqdnFailed: %v", err)
		}

		certPair, err := synth.SelfSignedCertificatePairFactory(&commonName, []tkValueObject.Fqdn{})
		if err != nil {
			t.Errorf("UnexpectedError: %v", err)
		}

		if certPair.Leaf.Subject.CommonName != "example.com" {
			t.Errorf("UnexpectedCommonName: '%s' vs 'example.com'", certPair.Leaf.Subject.CommonName)
		}
	})

	t.Run("WithAltNames", func(t *testing.T) {
		altName1, err := tkValueObject.NewFqdn("alt1.example.com")
		if err != nil {
			t.Fatalf("CreateFqdnFailed: %v", err)
		}
		altName2, err := tkValueObject.NewFqdn("alt2.example.com")
		if err != nil {
			t.Fatalf("CreateFqdnFailed: %v", err)
		}

		altNames := []tkValueObject.Fqdn{altName1, altName2}

		certPair, err := synth.SelfSignedCertificatePairFactory(nil, altNames)
		if err != nil {
			t.Errorf("UnexpectedError: %v", err)
		}

		expectedAltNames := []string{"alt1.example.com", "alt2.example.com"}
		if len(certPair.Leaf.DNSNames) != len(expectedAltNames) {
			t.Errorf("UnexpectedAltNamesCount: %d vs %d", len(certPair.Leaf.DNSNames), len(expectedAltNames))
		}

		for i, expected := range expectedAltNames {
			if i >= len(certPair.Leaf.DNSNames) || certPair.Leaf.DNSNames[i] != expected {
				t.Errorf("UnexpectedDNSName: '%s' vs '%s'", certPair.Leaf.DNSNames[i], expected)
			}
		}
	})

	t.Run("CertificateUsage", func(t *testing.T) {
		certPair, err := synth.SelfSignedCertificatePairFactory(nil, []tkValueObject.Fqdn{})
		if err != nil {
			t.Errorf("UnexpectedError: %v", err)
		}

		expectedKeyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		if certPair.Leaf.KeyUsage != expectedKeyUsage {
			t.Errorf("UnexpectedKeyUsage: %v vs %v", certPair.Leaf.KeyUsage, expectedKeyUsage)
		}

		if len(certPair.Leaf.ExtKeyUsage) != 1 || certPair.Leaf.ExtKeyUsage[0] != x509.ExtKeyUsageServerAuth {
			t.Errorf("UnexpectedExtKeyUsage: %v", certPair.Leaf.ExtKeyUsage)
		}
	})
}

func TestSelfSignedCertificatePairPemFactory(t *testing.T) {
	synth := &Synthesizer{}

	commonNameExample, _ := tkValueObject.NewFqdn("test.example.com")
	altName1, _ := tkValueObject.NewFqdn("alt1.test.com")
	altName2, _ := tkValueObject.NewFqdn("alt2.test.com")

	testCases := []struct {
		name               string
		commonNamePtr      *tkValueObject.Fqdn
		altNames           []tkValueObject.Fqdn
		expectedCommonName string
		expectedAltNames   []string
	}{
		{
			name:               "WithNilCommonNameAndEmptyAltNames",
			commonNamePtr:      nil,
			altNames:           []tkValueObject.Fqdn{},
			expectedCommonName: "localhost",
			expectedAltNames:   []string{},
		},
		{
			name:               "WithProvidedCommonName",
			commonNamePtr:      &commonNameExample,
			altNames:           []tkValueObject.Fqdn{},
			expectedCommonName: "test.example.com",
			expectedAltNames:   []string{},
		},
		{
			name:               "WithAltNames",
			commonNamePtr:      nil,
			altNames:           []tkValueObject.Fqdn{altName1, altName2},
			expectedCommonName: "localhost",
			expectedAltNames:   []string{"alt1.test.com", "alt2.test.com"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			certPem, keyPem, err := synth.SelfSignedCertificatePairPemFactory(testCase.commonNamePtr, testCase.altNames)
			if err != nil {
				t.Errorf("SelfSignedCertificatePairPemFactoryError: %v", err)
			}

			if !strings.HasPrefix(certPem, "-----BEGIN CERTIFICATE-----") {
				t.Error("CertPemHeaderInvalid")
			}
			if !strings.HasSuffix(certPem, "-----END CERTIFICATE-----\n") {
				t.Error("CertPemFooterInvalid")
			}

			if !strings.HasPrefix(keyPem, "-----BEGIN EC PRIVATE KEY-----") {
				t.Error("KeyPemHeaderInvalid")
			}
			if !strings.HasSuffix(keyPem, "-----END EC PRIVATE KEY-----\n") {
				t.Error("KeyPemFooterInvalid")
			}

			certificatePemBlock, _ := pem.Decode([]byte(certPem))
			if certificatePemBlock == nil || certificatePemBlock.Type != "CERTIFICATE" {
				t.Error("CertPemBlockDecodeFail")
			}
			parsedCert, err := x509.ParseCertificate(certificatePemBlock.Bytes)
			if err != nil {
				t.Errorf("CertParseFail: %v", err)
			}
			if parsedCert.Subject.CommonName != testCase.expectedCommonName {
				t.Errorf("CommonNameMismatch: Expected '%s', Got '%s'", testCase.expectedCommonName, parsedCert.Subject.CommonName)
			}
			if len(parsedCert.DNSNames) != len(testCase.expectedAltNames) {
				t.Errorf("AltNamesCountMismatch: Expected %d, Got %d", len(testCase.expectedAltNames), len(parsedCert.DNSNames))
			}
			for altNameIndex, expectedAltName := range testCase.expectedAltNames {
				if altNameIndex >= len(parsedCert.DNSNames) || parsedCert.DNSNames[altNameIndex] != expectedAltName {
					t.Errorf("AltNameMismatch: Expected '%s', Got '%s'", expectedAltName, parsedCert.DNSNames[altNameIndex])
				}
			}

			if len(certPem) == 0 {
				t.Error("CertPemEmpty")
			}
			if len(keyPem) == 0 {
				t.Error("KeyPemEmpty")
			}
		})
	}
}
