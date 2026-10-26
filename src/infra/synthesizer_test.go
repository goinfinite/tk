package tkInfra

import (
	"crypto/x509"
	"encoding/pem"
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestRandomIntegerGenerator(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("ValuesStayWithinInclusiveBounds", func(subtest *testing.T) {
		testCases := []struct {
			lowestValue  int
			highestValue int
		}{
			{3, 7},
			{-10, -5},
			{0, 0},
			{-2, 2},
			{100, 257},
			{math.MinInt, math.MaxInt},
		}

		for _, testCase := range testCases {
			for range 200 {
				randomValue := synth.RandomIntegerGenerator(
					testCase.lowestValue, testCase.highestValue,
				)

				valueIsInRange := randomValue >= testCase.lowestValue &&
					randomValue <= testCase.highestValue
				if !valueIsInRange {
					subtest.Errorf(
						"ValueOutsideBounds: %d [%d, %d]",
						randomValue, testCase.lowestValue,
						testCase.highestValue,
					)
				}
			}
		}
	})

	test.Run("EveryValueInSmallRangeIsReachable", func(subtest *testing.T) {
		lowestValue := 3
		highestValue := 7

		seenValues := map[int]bool{}
		for range 200 {
			seenValues[synth.RandomIntegerGenerator(
				lowestValue, highestValue,
			)] = true
		}

		for expectedValue := lowestValue; expectedValue <= highestValue; expectedValue++ {
			if !seenValues[expectedValue] {
				subtest.Errorf(
					"UnreachedValue: %d [%d, %d]",
					expectedValue, lowestValue, highestValue,
				)
			}
		}
	})

	test.Run("InvertedRangePanics", func(subtest *testing.T) {
		defer func() {
			recoveredValue := recover()
			if recoveredValue == nil {
				subtest.Fatal("ExpectedPanicForInvertedRange")
			}
			if recoveredValue != "RandomIntegerRangeInvalid" {
				subtest.Errorf("UnexpectedPanicValue: %v", recoveredValue)
			}
		}()

		synth.RandomIntegerGenerator(10, 5)
	})
}

func TestCharsetPresenceGuarantor(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("CharsetAlreadyPresent", func(subtest *testing.T) {
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
				subtest.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' [%s, %s]",
					string(actualOutput), string(testCase.expectedOutput), string(testCase.inputString),
					testCase.charset,
				)
			}
		}
	})

	test.Run("CharsetNotPresent", func(subtest *testing.T) {
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
				subtest.Errorf(
					"OutputUnchanged: Expected modification to include charset '%s' in '%s'",
					testCase.charset, string(testCase.inputString),
				)
			}

			if !strings.ContainsAny(string(actualOutput), testCase.charset) {
				subtest.Errorf(
					"CharsetNotAdded: Output '%s' does not contain any character from charset '%s'",
					string(actualOutput), testCase.charset,
				)
			}

			differentChars := 0
			for charIndex := 0; charIndex < len(testCase.inputString); charIndex++ {
				if actualOutput[charIndex] != testCase.inputString[charIndex] {
					differentChars++
				}
			}
			if differentChars != 1 {
				subtest.Errorf(
					"UnexpectedModificationCount: %d characters modified instead of 1 [%s, %s]",
					differentChars, string(testCase.inputString), testCase.charset,
				)
			}
		}
	})

	test.Run("EdgeCases", func(subtest *testing.T) {
		singleChar := []byte("a")
		charset := "b"
		actualOutput := synth.CharsetPresenceGuarantor(singleChar, charset)
		if string(actualOutput) != "b" {
			subtest.Errorf(
				"UnexpectedOutputValue: '%s' vs 'b' [single character case]",
				string(actualOutput),
			)
		}

		twoChars := []byte("ab")
		charset = "c"
		actualOutput = synth.CharsetPresenceGuarantor(twoChars, charset)
		if !strings.ContainsAny(string(actualOutput), charset) {
			subtest.Errorf(
				"CharsetNotAdded: Output '%s' does not contain any character from charset '%s'",
				string(actualOutput), charset,
			)
		}
	})

	test.Run("InputLargerThanOneByteRange", func(subtest *testing.T) {
		longInput := []byte(strings.Repeat("a", 300))
		actualOutput := synth.CharsetPresenceGuarantor(longInput, "9")
		if !strings.ContainsAny(string(actualOutput), "9") {
			subtest.Errorf("CharsetNotAddedTo300ByteInput")
		}
	})
}

func TestPasswordFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("PasswordLength", func(subtest *testing.T) {
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
				subtest.Errorf(
					"UnexpectedPasswordLength: '%d' vs '%d' [desired: %d, includeSymbols: %t]",
					len(password), testCase.desiredLength, testCase.desiredLength,
					testCase.shouldIncludeSymbols,
				)
			}
		}
	})

	test.Run("PasswordCharacteristics", func(subtest *testing.T) {
		password := synth.PasswordFactory(12, false)
		if !strings.ContainsAny(password, CharsetLowercaseLetters) {
			subtest.Errorf("MissingLowercaseLetters: '%s'", password)
		}

		if !strings.ContainsAny(password, CharsetUppercaseLetters) {
			subtest.Errorf("MissingUppercaseLetters: '%s'", password)
		}

		if !strings.ContainsAny(password, CharsetNumbers) {
			subtest.Errorf("MissingNumbers: '%s'", password)
		}

		if len(password) != 12 {
			subtest.Errorf("UnexpectedPasswordLength: '%d' vs '12'", len(password))
		}

		passwordWithSymbols := synth.PasswordFactory(12, true)
		if !strings.ContainsAny(passwordWithSymbols, CharsetSymbols) {
			subtest.Errorf("MissingSymbols: '%s'", passwordWithSymbols)
		}

		if len(passwordWithSymbols) != 12 {
			subtest.Errorf("UnexpectedPasswordLength: '%d' vs '12'", len(passwordWithSymbols))
		}
	})

	test.Run("ShortPasswordCharacteristics", func(subtest *testing.T) {
		shortPassword := synth.PasswordFactory(4, false)
		if len(shortPassword) != 4 {
			subtest.Errorf("UnexpectedPasswordLength: '%d' vs '4'", len(shortPassword))
		}

		veryShortPassword := synth.PasswordFactory(3, true)
		if len(veryShortPassword) != 3 {
			subtest.Errorf("UnexpectedPasswordLength: '%d' vs '3'", len(veryShortPassword))
		}
	})
}

func TestUsernameFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("UsernameGeneration", func(subtest *testing.T) {
		usernamesRegex := `^\w{1,256}$`
		re := regexp.MustCompile(usernamesRegex)

		for range 5 {
			username := synth.UsernameFactory()
			if !re.MatchString(username) {
				subtest.Errorf(
					"InvalidUsernameFormat: '%s' does not match regex '%s'",
					username, usernamesRegex,
				)
			}
		}
	})
}

func TestMailAddressFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("WithNilUsername", func(subtest *testing.T) {
		rawMailAddress := synth.MailAddressFactory(nil)
		_, err := tkValueObject.NewMailAddress(rawMailAddress)
		if err != nil {
			subtest.Errorf("InvalidMailAddress: '%s' is not a valid email address", rawMailAddress)
		}
	})

	test.Run("WithProvidedUsername", func(subtest *testing.T) {
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
				subtest.Errorf("InvalidMailAddress: '%s' is not a valid email address", rawMailAddress)
			}

			if !strings.HasPrefix(mailAddress.String(), username) {
				subtest.Errorf("MissingUsername: '%s' does not start with '%s'", mailAddress, username)
			}
		}
	})
}

func TestSelfSignedCertificatePairFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("WithNilCommonNameAndEmptyAltNames", func(subtest *testing.T) {
		certPair, err := synth.SelfSignedCertificatePairFactory(nil, []tkValueObject.Fqdn{})
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
		}

		if certPair.Leaf == nil {
			subtest.Error("CertificateLeafIsNil")
		}

		if certPair.Leaf.Subject.CommonName != "localhost" {
			subtest.Errorf("UnexpectedCommonName: '%s' vs 'localhost'", certPair.Leaf.Subject.CommonName)
		}

		if len(certPair.Leaf.DNSNames) != 0 {
			subtest.Errorf("UnexpectedAltNames: %v", certPair.Leaf.DNSNames)
		}

		if certPair.Leaf.Subject.Organization[0] != "Daystrom Institute" {
			subtest.Errorf("UnexpectedOrganization: '%s' vs 'Daystrom Institute'", certPair.Leaf.Subject.Organization[0])
		}

		if certPair.Leaf.SerialNumber == nil || certPair.Leaf.SerialNumber.Sign() <= 0 {
			subtest.Error("InvalidSerialNumber")
		}

		validFrom := certPair.Leaf.NotBefore
		validUntil := certPair.Leaf.NotAfter
		if validUntil.Sub(validFrom) != 365*24*time.Hour {
			subtest.Errorf("UnexpectedValidityPeriod: %v", validUntil.Sub(validFrom))
		}
	})

	test.Run("WithProvidedCommonName", func(subtest *testing.T) {
		commonName, err := tkValueObject.NewFqdn("example.com")
		if err != nil {
			subtest.Fatalf("CreateFqdnFailed: %v", err)
		}

		certPair, err := synth.SelfSignedCertificatePairFactory(&commonName, []tkValueObject.Fqdn{})
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
		}

		if certPair.Leaf.Subject.CommonName != "example.com" {
			subtest.Errorf("UnexpectedCommonName: '%s' vs 'example.com'", certPair.Leaf.Subject.CommonName)
		}
	})

	test.Run("WithAltNames", func(subtest *testing.T) {
		altName1, err := tkValueObject.NewFqdn("alt1.example.com")
		if err != nil {
			subtest.Fatalf("CreateFqdnFailed: %v", err)
		}
		altName2, err := tkValueObject.NewFqdn("alt2.example.com")
		if err != nil {
			subtest.Fatalf("CreateFqdnFailed: %v", err)
		}

		altNames := []tkValueObject.Fqdn{altName1, altName2}

		certPair, err := synth.SelfSignedCertificatePairFactory(nil, altNames)
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
		}

		expectedAltNames := []string{"alt1.example.com", "alt2.example.com"}
		if len(certPair.Leaf.DNSNames) != len(expectedAltNames) {
			subtest.Errorf(
				"UnexpectedAltNamesCount: %d vs %d",
				len(certPair.Leaf.DNSNames), len(expectedAltNames),
			)
		}

		for altNameIndex, expectedAltName := range expectedAltNames {
			if altNameIndex >= len(certPair.Leaf.DNSNames) {
				subtest.Errorf(
					"MissingDNSNameAtIndex: %d (expected '%s')",
					altNameIndex, expectedAltName,
				)
				continue
			}

			if certPair.Leaf.DNSNames[altNameIndex] != expectedAltName {
				subtest.Errorf(
					"UnexpectedDNSName: '%s' vs '%s'",
					certPair.Leaf.DNSNames[altNameIndex], expectedAltName,
				)
			}
		}
	})

	test.Run("CertificateUsage", func(subtest *testing.T) {
		certPair, err := synth.SelfSignedCertificatePairFactory(nil, []tkValueObject.Fqdn{})
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
		}

		expectedKeyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		if certPair.Leaf.KeyUsage != expectedKeyUsage {
			subtest.Errorf("UnexpectedKeyUsage: %v vs %v", certPair.Leaf.KeyUsage, expectedKeyUsage)
		}

		if len(certPair.Leaf.ExtKeyUsage) != 1 || certPair.Leaf.ExtKeyUsage[0] != x509.ExtKeyUsageServerAuth {
			subtest.Errorf("UnexpectedExtKeyUsage: %v", certPair.Leaf.ExtKeyUsage)
		}
	})
}

func TestSelfSignedCertificatePairPemFactory(test *testing.T) {
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
		test.Run(testCase.name, func(subtest *testing.T) {
			certPem, keyPem, err := synth.SelfSignedCertificatePairPemFactory(testCase.commonNamePtr, testCase.altNames)
			if err != nil {
				subtest.Errorf("SelfSignedCertificatePairPemFactoryError: %v", err)
			}

			if !strings.HasPrefix(certPem, "-----BEGIN CERTIFICATE-----") {
				subtest.Error("CertPemHeaderInvalid")
			}
			if !strings.HasSuffix(certPem, "-----END CERTIFICATE-----\n") {
				subtest.Error("CertPemFooterInvalid")
			}

			if !strings.HasPrefix(keyPem, "-----BEGIN EC PRIVATE KEY-----") {
				subtest.Error("KeyPemHeaderInvalid")
			}
			if !strings.HasSuffix(keyPem, "-----END EC PRIVATE KEY-----\n") {
				subtest.Error("KeyPemFooterInvalid")
			}

			certificatePemBlock, _ := pem.Decode([]byte(certPem))
			if certificatePemBlock == nil || certificatePemBlock.Type != "CERTIFICATE" {
				subtest.Error("CertPemBlockDecodeFail")
			}
			parsedCert, err := x509.ParseCertificate(certificatePemBlock.Bytes)
			if err != nil {
				subtest.Errorf("CertParseFail: %v", err)
			}
			if parsedCert.Subject.CommonName != testCase.expectedCommonName {
				subtest.Errorf(
					"CommonNameMismatch: Expected '%s', Got '%s'",
					testCase.expectedCommonName, parsedCert.Subject.CommonName,
				)
			}
			if len(parsedCert.DNSNames) != len(testCase.expectedAltNames) {
				subtest.Errorf(
					"AltNamesCountMismatch: Expected %d, Got %d",
					len(testCase.expectedAltNames), len(parsedCert.DNSNames),
				)
			}
			for altNameIndex, expectedAltName := range testCase.expectedAltNames {
				if altNameIndex >= len(parsedCert.DNSNames) {
					subtest.Errorf(
						"AltNameMissing: Expected '%s' at index %d",
						expectedAltName, altNameIndex,
					)
					continue
				}

				if parsedCert.DNSNames[altNameIndex] != expectedAltName {
					subtest.Errorf(
						"AltNameMismatch: Expected '%s', Got '%s'",
						expectedAltName, parsedCert.DNSNames[altNameIndex],
					)
				}
			}

			if len(certPem) == 0 {
				subtest.Error("CertPemEmpty")
			}
			if len(keyPem) == 0 {
				subtest.Error("KeyPemEmpty")
			}
		})
	}
}

func TestPrivateKeyPemFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("RSAPrivateKey", func(subtest *testing.T) {
		testCases := []struct {
			bitSize         int
			expectedBitSize int
		}{
			{0, 2048}, // default bit size
			{2048, 2048},
			{4096, 4096},
		}

		for _, testCase := range testCases {
			keyPem, err := synth.PrivateKeyPemFactory(PrivateKeySettings{
				Algorithm: tkValueObject.PrivateKeyAlgorithmRSA,
				BitSize:   testCase.bitSize,
			})
			if err != nil {
				subtest.Errorf("UnexpectedError: %v [bitSize: %d]", err, testCase.bitSize)
				continue
			}

			if !strings.HasPrefix(keyPem, "-----BEGIN RSA PRIVATE KEY-----") {
				subtest.Errorf("InvalidRSAKeyHeader: %s", keyPem[:min(50, len(keyPem))])
			}
			if !strings.HasSuffix(keyPem, "-----END RSA PRIVATE KEY-----\n") {
				subtest.Errorf("InvalidRSAKeyFooter")
			}

			decodedPemBlock, _ := pem.Decode([]byte(keyPem))
			if decodedPemBlock == nil || decodedPemBlock.Type != "RSA PRIVATE KEY" {
				subtest.Errorf("DecodePEMFailed")
				continue
			}

			parsedRsaKey, err := x509.ParsePKCS1PrivateKey(decodedPemBlock.Bytes)
			if err != nil {
				subtest.Fatalf(
					"FailedToParseRSAKey: %v [bitSize: %d]",
					err, testCase.bitSize,
				)
			}

			actualBitSize := parsedRsaKey.N.BitLen()
			if actualBitSize != testCase.expectedBitSize {
				subtest.Errorf(
					"UnexpectedBitSize: %d vs %d [requested: %d]",
					actualBitSize, testCase.expectedBitSize, testCase.bitSize,
				)
			}
		}
	})

	test.Run("ECDSAPrivateKey", func(subtest *testing.T) {
		testCases := []struct {
			bitSize       int
			expectedCurve string
		}{
			{0, "P-256"}, // default curve
			{256, "P-256"},
			{384, "P-384"},
			{521, "P-521"},
		}

		for _, testCase := range testCases {
			keyPem, err := synth.PrivateKeyPemFactory(PrivateKeySettings{
				Algorithm: tkValueObject.PrivateKeyAlgorithmECDSA,
				BitSize:   testCase.bitSize,
			})
			if err != nil {
				subtest.Errorf("UnexpectedError: %v [bitSize: %d]", err, testCase.bitSize)
				continue
			}

			if !strings.HasPrefix(keyPem, "-----BEGIN EC PRIVATE KEY-----") {
				subtest.Errorf("InvalidECKeyHeader")
			}
			if !strings.HasSuffix(keyPem, "-----END EC PRIVATE KEY-----\n") {
				subtest.Errorf("InvalidECKeyFooter")
			}

			decodedPemBlock, _ := pem.Decode([]byte(keyPem))
			if decodedPemBlock == nil || decodedPemBlock.Type != "EC PRIVATE KEY" {
				subtest.Errorf("DecodePEMFailed")
				continue
			}

			parsedEcdsaKey, err := x509.ParseECPrivateKey(decodedPemBlock.Bytes)
			if err != nil {
				subtest.Fatalf(
					"ParseECDSAKeyFailed: %v [bitSize: %d]",
					err, testCase.bitSize,
				)
			}

			actualCurveName := parsedEcdsaKey.Curve.Params().Name
			if actualCurveName != testCase.expectedCurve {
				subtest.Errorf(
					"UnexpectedCurve: %s vs %s [requested bitSize: %d]",
					actualCurveName, testCase.expectedCurve, testCase.bitSize,
				)
			}
		}
	})

	test.Run("DSAPrivateKey", func(subtest *testing.T) {
		keyPem, err := synth.PrivateKeyPemFactory(PrivateKeySettings{
			Algorithm: tkValueObject.PrivateKeyAlgorithmDSA,
		})
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
			return
		}

		if !strings.HasPrefix(keyPem, "-----BEGIN DSA PRIVATE KEY-----") {
			subtest.Errorf("InvalidDSAKeyHeader")
		}
		if !strings.HasSuffix(keyPem, "-----END DSA PRIVATE KEY-----\n") {
			subtest.Errorf("InvalidDSAKeyFooter")
		}

		decodedPemBlock, _ := pem.Decode([]byte(keyPem))
		if decodedPemBlock == nil || decodedPemBlock.Type != "DSA PRIVATE KEY" {
			subtest.Errorf("DecodePEMFailed")
		}
	})

	test.Run("Ed25519PrivateKey", func(subtest *testing.T) {
		keyPem, err := synth.PrivateKeyPemFactory(PrivateKeySettings{
			Algorithm: tkValueObject.PrivateKeyAlgorithmEd25519,
		})
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
			return
		}

		if !strings.HasPrefix(keyPem, "-----BEGIN PRIVATE KEY-----") {
			subtest.Errorf("InvalidEd25519KeyHeader")
		}
		if !strings.HasSuffix(keyPem, "-----END PRIVATE KEY-----\n") {
			subtest.Errorf("InvalidEd25519KeyFooter")
		}

		decodedPemBlock, _ := pem.Decode([]byte(keyPem))
		if decodedPemBlock == nil || decodedPemBlock.Type != "PRIVATE KEY" {
			subtest.Errorf("DecodePEMFailed")
		}
	})

	test.Run("DefaultAlgorithm", func(subtest *testing.T) {
		keyPem, err := synth.PrivateKeyPemFactory(PrivateKeySettings{})
		if err != nil {
			subtest.Errorf("UnexpectedError: %v", err)
			return
		}

		// Default should be ECDSA
		if !strings.HasPrefix(keyPem, "-----BEGIN EC PRIVATE KEY-----") {
			subtest.Errorf("DefaultShouldBeECDSA")
		}
	})

	test.Run("UnsupportedAlgorithm", func(subtest *testing.T) {
		_, err := synth.PrivateKeyPemFactory(PrivateKeySettings{
			Algorithm: "INVALID",
		})
		if err == nil {
			subtest.Errorf("MissingExpectedError: unsupported algorithm should fail")
		}
		if err.Error() != "UnsupportedPrivateKeyAlgorithm" {
			subtest.Errorf("UnexpectedErrorMessage: %s", err.Error())
		}
	})

	test.Run("InvalidECDSABitSize", func(subtest *testing.T) {
		invalidBitSizes := []int{128, 192, 224, 512, 1024}

		for _, bitSize := range invalidBitSizes {
			_, err := synth.PrivateKeyPemFactory(PrivateKeySettings{
				Algorithm: tkValueObject.PrivateKeyAlgorithmECDSA,
				BitSize:   bitSize,
			})
			if err == nil {
				subtest.Errorf(
					"MissingExpectedError: invalid ECDSA bitSize %d should fail",
					bitSize,
				)
			}
			if err != nil && err.Error() != "InvalidECDSABitSize" {
				subtest.Errorf(
					"UnexpectedErrorMessage for bitSize %d: %s",
					bitSize, err.Error(),
				)
			}
		}
	})
}

func TestCertificatePemFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("DefaultCertificate", func(subtest *testing.T) {
		certPem, keyPem, err := synth.CertificatePemFactory(CertificateSettings{})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if !strings.HasPrefix(certPem, "-----BEGIN CERTIFICATE-----") {
			subtest.Error("InvalidCertHeader")
		}
		if !strings.HasSuffix(certPem, "-----END CERTIFICATE-----\n") {
			subtest.Error("InvalidCertFooter")
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		if decodedPemBlock == nil || decodedPemBlock.Type != "CERTIFICATE" {
			subtest.Error("DecodeCertPEMFailed")
		}

		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if cert.Subject.CommonName != "localhost" {
			subtest.Errorf("UnexpectedCommonName: %s", cert.Subject.CommonName)
		}
		if cert.IsCA {
			subtest.Error("DefaultCertShouldNotBeCA")
		}

		if !strings.HasPrefix(keyPem, "-----BEGIN EC PRIVATE KEY-----") {
			subtest.Error("InvalidKeyHeader")
		}
	})

	test.Run("WithCommonName", func(subtest *testing.T) {
		commonName, _ := tkValueObject.NewFqdn("test.example.com")
		certPem, _, err := synth.CertificatePemFactory(CertificateSettings{
			CommonName: &commonName,
		})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if cert.Subject.CommonName != "test.example.com" {
			subtest.Errorf("UnexpectedCommonName: %s", cert.Subject.CommonName)
		}
	})

	test.Run("WithAltNames", func(subtest *testing.T) {
		altName1, _ := tkValueObject.NewFqdn("alt1.example.com")
		altName2, _ := tkValueObject.NewFqdn("alt2.example.com")

		certPem, _, err := synth.CertificatePemFactory(CertificateSettings{
			AltNames: []tkValueObject.Fqdn{altName1, altName2},
		})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if len(cert.DNSNames) != 2 {
			subtest.Errorf("UnexpectedAltNamesCount: %d", len(cert.DNSNames))
		}
		if cert.DNSNames[0] != "alt1.example.com" || cert.DNSNames[1] != "alt2.example.com" {
			subtest.Errorf("UnexpectedAltNames: %v", cert.DNSNames)
		}
	})
}

func TestCACertificatePemFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("DefaultCACertificate", func(subtest *testing.T) {
		certPem, keyPem, err := synth.CACertificatePemFactory(CertificateSettings{})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if !strings.HasPrefix(certPem, "-----BEGIN CERTIFICATE-----") {
			subtest.Error("InvalidCertHeader")
		}
		if !strings.HasPrefix(keyPem, "-----BEGIN EC PRIVATE KEY-----") {
			subtest.Error("InvalidKeyHeader")
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if !cert.IsCA {
			subtest.Error("CertShouldBeCA")
		}
		if cert.Subject.CommonName != "Test CA" {
			subtest.Errorf("UnexpectedCommonName: %s", cert.Subject.CommonName)
		}

		expectedKeyUsage := x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		if cert.KeyUsage != expectedKeyUsage {
			subtest.Errorf("UnexpectedKeyUsage: %v", cert.KeyUsage)
		}
	})

	test.Run("CACertificateWithMaxPathLen", func(subtest *testing.T) {
		maxPathLen := 2
		certPem, _, err := synth.CACertificatePemFactory(CertificateSettings{
			MaxPathLengthPtr: &maxPathLen,
		})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if cert.MaxPathLen != 2 {
			subtest.Errorf("UnexpectedMaxPathLen: %d", cert.MaxPathLen)
		}
	})

	test.Run("CACertificateWithMaxPathLenZero", func(subtest *testing.T) {
		certPem, _, err := synth.CACertificatePemFactory(CertificateSettings{
			HasMaxPathLengthZero: true,
		})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if cert.MaxPathLen != 0 {
			subtest.Errorf("UnexpectedMaxPathLen: %d", cert.MaxPathLen)
		}
		if !cert.MaxPathLenZero {
			subtest.Error("MaxPathLenZeroShouldBeTrue")
		}
	})

	test.Run("CACertificateWithCustomCommonName", func(subtest *testing.T) {
		commonName, _ := tkValueObject.NewFqdn("custom-ca.example.com")
		certPem, _, err := synth.CACertificatePemFactory(CertificateSettings{
			CommonName: &commonName,
		})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if cert.Subject.CommonName != "custom-ca.example.com" {
			subtest.Errorf("UnexpectedCommonName: %s", cert.Subject.CommonName)
		}
	})

	test.Run("CACertificateSetsIsCATrue", func(subtest *testing.T) {
		// Test that CACertificatePemFactory always sets IsCA=true
		// even if the settings have IsCA=false
		certPem, _, err := synth.CACertificatePemFactory(CertificateSettings{
			IsCA: false, // This should be overridden
		})
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		decodedPemBlock, _ := pem.Decode([]byte(certPem))
		cert, err := x509.ParseCertificate(decodedPemBlock.Bytes)
		if err != nil {
			subtest.Fatalf("ParseCertificateFailed: %v", err)
		}

		if !cert.IsCA {
			subtest.Error("CACertificatePemFactoryShouldAlwaysSetIsCATrue")
		}
	})
}

func TestRandomStringFactory(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("LengthAndCharsetMembership", func(subtest *testing.T) {
		testCases := []struct {
			length  int
			charset string
		}{
			{1, CharsetLowercaseLetters},
			{8, CharsetLowercaseLetters + CharsetUppercaseLetters + CharsetNumbers},
			{64, CharsetSymbols},
			{256, CharsetNumbers},
		}

		for _, testCase := range testCases {
			generated := synth.randomStringFactory(
				testCase.length,
				testCase.charset,
			)

			if len(generated) != testCase.length {
				subtest.Errorf(
					"LengthMismatch: expected %d, got %d",
					testCase.length, len(generated),
				)
			}

			for _, generatedChar := range generated {
				charsetContainsChar := strings.ContainsRune(
					testCase.charset, generatedChar,
				)
				if !charsetContainsChar {
					subtest.Errorf(
						"CharOutsideCharset: '%c' not in %s",
						generatedChar, testCase.charset,
					)
				}
			}
		}
	})

	test.Run("NonPositiveLength", func(subtest *testing.T) {
		for _, invalidLength := range []int{0, -1, -128} {
			generated := synth.randomStringFactory(
				invalidLength,
				CharsetLowercaseLetters,
			)
			if generated != "" {
				subtest.Errorf("UnexpectedValueForLength %d: '%s'", invalidLength, generated)
			}
		}
	})

	test.Run("EmptyCharsetYieldsEmptyString", func(subtest *testing.T) {
		generated := synth.randomStringFactory(8, "")
		if generated != "" {
			subtest.Errorf("UnexpectedValueForEmptyCharset: '%s'", generated)
		}
	})

	test.Run("DistinctAcrossCalls", func(subtest *testing.T) {
		alphanumeric := CharsetLowercaseLetters +
			CharsetUppercaseLetters + CharsetNumbers

		first := synth.randomStringFactory(32, alphanumeric)
		second := synth.randomStringFactory(32, alphanumeric)

		if first == second {
			subtest.Errorf("IdenticalValuesAcrossCalls: %s", first)
		}
	})

	test.Run("NoPerCharacterBias", func(subtest *testing.T) {
		charset := "abcd"
		samplesTotal := 400
		sampleLength := 250

		bucketCounts := map[rune]int{}
		for range samplesTotal {
			generated := synth.randomStringFactory(sampleLength, charset)
			for _, generatedChar := range generated {
				bucketCounts[generatedChar]++
			}
		}

		drawsTotal := samplesTotal * sampleLength
		expectedPerChar := drawsTotal / len(charset)
		tolerance := expectedPerChar / 10
		for _, charsetChar := range charset {
			charCount := bucketCounts[charsetChar]
			countInsideTolerance := charCount >= expectedPerChar-tolerance &&
				charCount <= expectedPerChar+tolerance
			if !countInsideTolerance {
				subtest.Errorf(
					"CharacterCountOutsideTolerance: '%c' got %d, expected %d±%d",
					charsetChar, charCount, expectedPerChar, tolerance,
				)
			}
		}
	})
}

func TestShuffledIndicesGenerator(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("ContainsEveryIndexExactlyOnce", func(subtest *testing.T) {
		for _, size := range []int{0, 1, 2, 3, 8, 64} {
			shuffledIndices := synth.shuffledIndicesGenerator(size)

			if len(shuffledIndices) != size {
				subtest.Errorf(
					"UnexpectedIndicesCount: expected %d, got %d",
					size, len(shuffledIndices),
				)
				continue
			}

			seenIndices := make(map[int]bool, size)
			for _, index := range shuffledIndices {
				indexIsInRange := index >= 0 && index < size
				if !indexIsInRange {
					subtest.Errorf("IndexOutsideRange: %d [size %d]", index, size)
				}
				if seenIndices[index] {
					subtest.Errorf("RepeatedIndex: %d [size %d]", index, size)
				}
				seenIndices[index] = true
			}
		}
	})

	test.Run("DistinctAcrossCalls", func(subtest *testing.T) {
		size := 16
		first := synth.shuffledIndicesGenerator(size)
		second := synth.shuffledIndicesGenerator(size)
		if slices.Equal(first, second) {
			subtest.Errorf("IdenticalOrderingsAcrossCalls: %v", first)
		}
	})
}

func TestPasswordFactoryClassGuarantees(test *testing.T) {
	synth := &Synthesizer{}

	test.Run("EveryClassPresentAboveMinimumLength", func(subtest *testing.T) {
		for range 20 {
			password := synth.PasswordFactory(16, true)

			if len(password) != 16 {
				subtest.Errorf("UnexpectedPasswordLength: %d", len(password))
			}
			for _, requiredCharset := range []string{
				CharsetLowercaseLetters,
				CharsetUppercaseLetters,
				CharsetNumbers,
				CharsetSymbols,
			} {
				if !strings.ContainsAny(password, requiredCharset) {
					subtest.Errorf("RequiredCharsetMissing: %s from %s", requiredCharset, password)
				}
			}
		}
	})

	test.Run("AlphanumericOnlyWhenSymbolsExcluded", func(subtest *testing.T) {
		alphanumeric := CharsetLowercaseLetters +
			CharsetUppercaseLetters + CharsetNumbers
		for range 20 {
			password := synth.PasswordFactory(12, false)
			if strings.ContainsAny(password, CharsetSymbols) {
				subtest.Errorf("UnexpectedSymbol: %s", password)
			}
			if !strings.ContainsAny(password, alphanumeric) {
				subtest.Errorf("UnexpectedCharacter: %s", password)
			}
		}
	})

	test.Run("DistinctAcrossCalls", func(subtest *testing.T) {
		first := synth.PasswordFactory(16, true)
		second := synth.PasswordFactory(16, true)
		if first == second {
			subtest.Errorf("IdenticalPasswordsAcrossCalls")
		}
	})

	test.Run("LengthAboveOneByteRange", func(subtest *testing.T) {
		password := synth.PasswordFactory(257, true)
		if len(password) != 257 {
			subtest.Errorf("UnexpectedPasswordLength: %d", len(password))
		}
		for _, requiredCharset := range []string{
			CharsetLowercaseLetters,
			CharsetUppercaseLetters,
			CharsetNumbers,
			CharsetSymbols,
		} {
			if !strings.ContainsAny(password, requiredCharset) {
				subtest.Errorf(
					"RequiredCharsetMissing: %s from 257-char password",
					requiredCharset,
				)
			}
		}
	})
}
