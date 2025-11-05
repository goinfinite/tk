package tkInfra

import (
	"os"
	"path/filepath"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestReadThrough_CertPairFilePathsReader(t *testing.T) {
	readThrough := &ReadThrough{}
	fileClerk := FileClerk{}

	t.Run("BothEnvVarsSetAndValid", func(t *testing.T) {
		tempDir := t.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")
		keyPath := filepath.Join(tempDir, "key.pem")

		os.Setenv("CERTIFICATE_PAIR_CERT_PATH", certPath)
		os.Setenv("CERTIFICATE_PAIR_KEY_PATH", keyPath)
		defer func() {
			os.Unsetenv("CERTIFICATE_PAIR_CERT_PATH")
			os.Unsetenv("CERTIFICATE_PAIR_KEY_PATH")
		}()

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(certPath, false)
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(keyPath, false)

		if actualCertPath.String() != expectedCertPath.String() {
			t.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			t.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}
	})

	t.Run("CertEnvVarInvalid", func(t *testing.T) {
		os.Setenv("CERTIFICATE_PAIR_CERT_PATH", "relative<path")
		os.Unsetenv("CERTIFICATE_PAIR_KEY_PATH")
		defer func() {
			os.Unsetenv("CERTIFICATE_PAIR_CERT_PATH")
		}()

		_, _, err := readThrough.CertPairFilePathsReader()
		if err == nil {
			t.Errorf("MissingExpectedError: InvalidCertificatePairCertPath")
		}
		if err.Error() != "InvalidCertificatePairCertPath" {
			t.Errorf("ErrorMismatch: Expected 'InvalidCertificatePairCertPath', Got '%s'", err.Error())
		}
	})

	t.Run("KeyEnvVarInvalid", func(t *testing.T) {
		tempDir := t.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")

		os.Setenv("CERTIFICATE_PAIR_CERT_PATH", certPath)
		os.Setenv("CERTIFICATE_PAIR_KEY_PATH", "relative<key")
		defer func() {
			os.Unsetenv("CERTIFICATE_PAIR_CERT_PATH")
			os.Unsetenv("CERTIFICATE_PAIR_KEY_PATH")
		}()

		_, _, err := readThrough.CertPairFilePathsReader()
		if err == nil {
			t.Errorf("MissingExpectedError: InvalidCertificatePairKeyPath")
		}
		if err.Error() != "InvalidCertificatePairKeyPath" {
			t.Errorf("ErrorMismatch: Expected 'InvalidCertificatePairKeyPath', Got '%s'", err.Error())
		}
	})

	t.Run("CertEnvVarSetKeyNotSet", func(t *testing.T) {
		tempDir := t.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")

		os.Setenv("CERTIFICATE_PAIR_CERT_PATH", certPath)
		os.Unsetenv("CERTIFICATE_PAIR_KEY_PATH")
		defer func() {
			os.Unsetenv("CERTIFICATE_PAIR_CERT_PATH")
		}()

		originalDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(originalDir)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPathStr := filepath.Join(tempDir, "pki", "cert.pem")
		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedCertPathStr, false)
		expectedKeyPathStr := filepath.Join(tempDir, "pki", "key.pem")
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedKeyPathStr, false)

		if actualCertPath.String() != expectedCertPath.String() {
			t.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			t.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}

		if !fileClerk.FileExists(expectedCertPathStr) {
			t.Errorf("CertFileNotCreated")
		}
		if !fileClerk.FileExists(expectedKeyPathStr) {
			t.Errorf("KeyFileNotCreated")
		}
	})

	t.Run("KeyEnvVarSetCertNotSet", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := filepath.Join(tempDir, "key.pem")

		os.Unsetenv("CERTIFICATE_PAIR_CERT_PATH")
		os.Setenv("CERTIFICATE_PAIR_KEY_PATH", keyPath)
		defer func() {
			os.Unsetenv("CERTIFICATE_PAIR_KEY_PATH")
		}()

		originalDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(originalDir)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPathStr := filepath.Join(tempDir, "pki", "cert.pem")
		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedCertPathStr, false)
		expectedKeyPathStr := filepath.Join(tempDir, "pki", "key.pem")
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedKeyPathStr, false)

		if actualCertPath.String() != expectedCertPath.String() {
			t.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			t.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}

		if !fileClerk.FileExists(expectedCertPathStr) {
			t.Errorf("CertFileNotCreated")
		}
		if !fileClerk.FileExists(expectedKeyPathStr) {
			t.Errorf("KeyFileNotCreated")
		}
	})

	t.Run("NeitherEnvVarSet", func(t *testing.T) {
		os.Unsetenv("CERTIFICATE_PAIR_CERT_PATH")
		os.Unsetenv("CERTIFICATE_PAIR_KEY_PATH")

		tempDir := t.TempDir()

		originalDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(originalDir)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPathStr := filepath.Join(tempDir, "pki", "cert.pem")
		expectedKeyPathStr := filepath.Join(tempDir, "pki", "key.pem")
		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedCertPathStr, false)
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedKeyPathStr, false)

		if actualCertPath.String() != expectedCertPath.String() {
			t.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			t.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}

		if !fileClerk.FileExists(expectedCertPathStr) {
			t.Errorf("CertFileNotCreated")
		}
		if !fileClerk.FileExists(expectedKeyPathStr) {
			t.Errorf("KeyFileNotCreated")
		}
	})
}
