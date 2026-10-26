package tkValueObject

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"
)

func TestNewX509PublicKeySize(test *testing.T) {
	test.Run("ValidPublicKeySize", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PublicKeySize
			expectError    bool
		}{
			{256, X509PublicKeySize(256), false},
			{384, X509PublicKeySize(384), false},
			{521, X509PublicKeySize(521), false},
			{1024, X509PublicKeySize(1024), false},
			{2048, X509PublicKeySize(2048), false},
			{3072, X509PublicKeySize(3072), false},
			{4096, X509PublicKeySize(4096), false},
			{8192, X509PublicKeySize(8192), false},
			{"2048", X509PublicKeySize(2048), false},
			{"4096", X509PublicKeySize(4096), false},
			{uint16(2048), X509PublicKeySize(2048), false},
			{2048.0, X509PublicKeySize(2048), false},
			{0, X509PublicKeySize(0), true},
			{128, X509PublicKeySize(0), true},
			{512, X509PublicKeySize(0), true},
			{1536, X509PublicKeySize(0), true},
			{16384, X509PublicKeySize(0), true},
			{2048.5, X509PublicKeySize(0), true},
			{"invalid", X509PublicKeySize(0), true},
			{-1, X509PublicKeySize(0), true},
			{nil, X509PublicKeySize(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PublicKeySize(testCase.inputValue)

			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				subtest.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("Uint16Method", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509PublicKeySize
			expectedOutput uint16
		}{
			{X509PublicKeySize(2048), 2048},
			{X509PublicKeySize(4096), 4096},
			{X509PublicKeySize(256), 256},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint16()

			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509PublicKeySize
			expectedOutput string
		}{
			{X509PublicKeySize(2048), "2048"},
			{X509PublicKeySize(4096), "4096"},
			{X509PublicKeySize(256), "256"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}

func TestNewX509PublicKeySizeFromStdlib(test *testing.T) {
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		test.Fatalf("RSAKeyGenerationFailed: %s", err.Error())
	}
	rsaPublicKey := &rsaPrivateKey.PublicKey

	ecdsaPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		test.Fatalf("ECDSAKeyGenerationFailed: %s", err.Error())
	}
	ecdsaPublicKey := &ecdsaPrivateKey.PublicKey

	hugeRsaPublicKey := &rsa.PublicKey{N: new(big.Int).Lsh(big.NewInt(1), 67583)}
	hugeEcdsaPublicKey := &ecdsa.PublicKey{
		Curve: &elliptic.CurveParams{BitSize: 67584},
	}

	testCaseStructs := []struct {
		inputValue     any
		expectedOutput X509PublicKeySize
		expectError    bool
	}{
		{rsaPublicKey, X509PublicKeySize(2048), false},
		{ecdsaPublicKey, X509PublicKeySize(256), false},
		{hugeRsaPublicKey, X509PublicKeySize(0), true},
		{hugeEcdsaPublicKey, X509PublicKeySize(0), true},
		{&rsa.PublicKey{}, X509PublicKeySize(0), true},
		{&ecdsa.PublicKey{}, X509PublicKeySize(0), true},
		{(*rsa.PublicKey)(nil), X509PublicKeySize(0), true},
		{(*ecdsa.PublicKey)(nil), X509PublicKeySize(0), true},
		{
			&ecdsa.PublicKey{Curve: (*elliptic.CurveParams)(nil)},
			X509PublicKeySize(0),
			true,
		},
		{"not a valid public key", X509PublicKeySize(0), true},
		{12345, X509PublicKeySize(0), true},
		{nil, X509PublicKeySize(0), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewX509PublicKeySizeFromStdlib(testCase.inputValue)

		if testCase.expectError && err == nil {
			test.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
		}

		if !testCase.expectError && err != nil {
			test.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
		}

		if !testCase.expectError &&
			actualOutput != testCase.expectedOutput {
			test.Errorf(
				"UnexpectedOutputValue: '%v' vs '%v' [%v]",
				actualOutput, testCase.expectedOutput, testCase.inputValue,
			)
		}
	}
}
