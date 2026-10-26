package tkInfra

import (
	"os"
	"path/filepath"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestReadThrough_CertPairFilePathsReader(test *testing.T) {
	readThrough := &ReadThrough{}
	fileClerk := FileClerk{}

	test.Run("BothEnvVarsSetAndValid", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")
		keyPath := filepath.Join(tempDir, "key.pem")

		subtest.Setenv("CERTIFICATE_PAIR_CERT_PATH", certPath)
		subtest.Setenv("CERTIFICATE_PAIR_KEY_PATH", keyPath)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(certPath, false)
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(keyPath, false)

		if actualCertPath.String() != expectedCertPath.String() {
			subtest.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			subtest.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}
	})

	test.Run("CertEnvVarInvalid", func(subtest *testing.T) {
		subtest.Setenv("CERTIFICATE_PAIR_CERT_PATH", "relative<path")
		subtest.Setenv("CERTIFICATE_PAIR_KEY_PATH", "")

		_, _, err := readThrough.CertPairFilePathsReader()
		if err == nil {
			subtest.Errorf("MissingExpectedError: InvalidCertificatePairCertPath")
		}
		if err.Error() != "InvalidCertificatePairCertPath" {
			subtest.Errorf("ErrorMismatch: Expected 'InvalidCertificatePairCertPath', Got '%s'", err.Error())
		}
	})

	test.Run("KeyEnvVarInvalid", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")

		subtest.Setenv("CERTIFICATE_PAIR_CERT_PATH", certPath)
		subtest.Setenv("CERTIFICATE_PAIR_KEY_PATH", "relative<key")

		_, _, err := readThrough.CertPairFilePathsReader()
		if err == nil {
			subtest.Errorf("MissingExpectedError: InvalidCertificatePairKeyPath")
		}
		if err.Error() != "InvalidCertificatePairKeyPath" {
			subtest.Errorf("ErrorMismatch: Expected 'InvalidCertificatePairKeyPath', Got '%s'", err.Error())
		}
	})

	test.Run("CertEnvVarSetKeyNotSet", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		certPath := filepath.Join(tempDir, "cert.pem")

		subtest.Setenv("CERTIFICATE_PAIR_CERT_PATH", certPath)
		subtest.Setenv("CERTIFICATE_PAIR_KEY_PATH", "")
		subtest.Chdir(tempDir)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPathStr := filepath.Join(tempDir, "pki", "cert.pem")
		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedCertPathStr, false)
		expectedKeyPathStr := filepath.Join(tempDir, "pki", "key.pem")
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedKeyPathStr, false)

		if actualCertPath.String() != expectedCertPath.String() {
			subtest.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			subtest.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}

		if !fileClerk.FileExists(expectedCertPathStr) {
			subtest.Errorf("CertFileNotCreated")
		}
		if !fileClerk.FileExists(expectedKeyPathStr) {
			subtest.Errorf("KeyFileNotCreated")
		}
	})

	test.Run("KeyEnvVarSetCertNotSet", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		keyPath := filepath.Join(tempDir, "key.pem")

		subtest.Setenv("CERTIFICATE_PAIR_CERT_PATH", "")
		subtest.Setenv("CERTIFICATE_PAIR_KEY_PATH", keyPath)
		subtest.Chdir(tempDir)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPathStr := filepath.Join(tempDir, "pki", "cert.pem")
		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedCertPathStr, false)
		expectedKeyPathStr := filepath.Join(tempDir, "pki", "key.pem")
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedKeyPathStr, false)

		if actualCertPath.String() != expectedCertPath.String() {
			subtest.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			subtest.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}

		if !fileClerk.FileExists(expectedCertPathStr) {
			subtest.Errorf("CertFileNotCreated")
		}
		if !fileClerk.FileExists(expectedKeyPathStr) {
			subtest.Errorf("KeyFileNotCreated")
		}
	})

	test.Run("NeitherEnvVarSet", func(subtest *testing.T) {
		subtest.Setenv("CERTIFICATE_PAIR_CERT_PATH", "")
		subtest.Setenv("CERTIFICATE_PAIR_KEY_PATH", "")

		tempDir := subtest.TempDir()
		subtest.Chdir(tempDir)

		actualCertPath, actualKeyPath, err := readThrough.CertPairFilePathsReader()
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		}

		expectedCertPathStr := filepath.Join(tempDir, "pki", "cert.pem")
		expectedKeyPathStr := filepath.Join(tempDir, "pki", "key.pem")
		expectedCertPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedCertPathStr, false)
		expectedKeyPath, _ := tkValueObject.NewUnixAbsoluteFilePath(expectedKeyPathStr, false)

		if actualCertPath.String() != expectedCertPath.String() {
			subtest.Errorf("CertPathMismatch: Expected '%s', Got '%s'", expectedCertPath.String(), actualCertPath.String())
		}
		if actualKeyPath.String() != expectedKeyPath.String() {
			subtest.Errorf("KeyPathMismatch: Expected '%s', Got '%s'", expectedKeyPath.String(), actualKeyPath.String())
		}

		if !fileClerk.FileExists(expectedCertPathStr) {
			subtest.Errorf("CertFileNotCreated")
		}
		if !fileClerk.FileExists(expectedKeyPathStr) {
			subtest.Errorf("KeyFileNotCreated")
		}

		certFileInfo, certStatErr := os.Stat(expectedCertPathStr)
		if certStatErr != nil {
			subtest.Fatalf("CertStatFailed: %v", certStatErr)
		}
		if certFileInfo.Mode().Perm() != 0644 {
			subtest.Errorf("CertPermissionsMismatch: %v", certFileInfo.Mode().Perm())
		}

		keyFileInfo, keyStatErr := os.Stat(expectedKeyPathStr)
		if keyStatErr != nil {
			subtest.Fatalf("KeyStatFailed: %v", keyStatErr)
		}
		if keyFileInfo.Mode().Perm() != 0600 {
			subtest.Errorf("KeyPermissionsMismatch: %v", keyFileInfo.Mode().Perm())
		}
	})
}
