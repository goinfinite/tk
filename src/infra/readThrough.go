package tkInfra

import (
	"errors"
	"os"
	"path/filepath"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	ReadThroughCertPairCertPathEnvVarName string = "CERTIFICATE_PAIR_CERT_PATH"
	ReadThroughCertPairKeyPathEnvVarName  string = "CERTIFICATE_PAIR_KEY_PATH"
	ReadThroughPkiDirEnvVarName           string = "PKI_DIR"
)

// Provides methods for reading information that when not found, are generated on the fly.
type ReadThrough struct {
}

// Attempts to retrieve the certificate pair file paths from the environment variables
// "CERTIFICATE_PAIR_CERT_PATH" and "CERTIFICATE_PAIR_KEY_PATH", otherwise generates a
// self-signed certificate pair on local 'pki' directory (or the directory specified by the
// environment variable "PKI_DIR") and returns the absolute paths to the generated files.
func (rt *ReadThrough) CertPairFilePathsReader() (
	certPath tkValueObject.UnixAbsoluteFilePath,
	keyPath tkValueObject.UnixAbsoluteFilePath,
	err error,
) {
	allowUnsafeFilePathChars := false

	isCertPathValid := false
	rawCertPath := os.Getenv(ReadThroughCertPairCertPathEnvVarName)
	if rawCertPath != "" {
		certPath, err = tkValueObject.NewUnixAbsoluteFilePath(rawCertPath, allowUnsafeFilePathChars)
		if err != nil {
			return certPath, keyPath, errors.New("InvalidCertificatePairCertPath")
		}
		isCertPathValid = true
	}

	isKeyPathValid := false
	rawKeyPath := os.Getenv(ReadThroughCertPairKeyPathEnvVarName)
	if rawKeyPath != "" {
		keyPath, err = tkValueObject.NewUnixAbsoluteFilePath(rawKeyPath, allowUnsafeFilePathChars)
		if err != nil {
			return certPath, keyPath, errors.New("InvalidCertificatePairKeyPath")
		}
		isKeyPathValid = true
	}

	if isCertPathValid && isKeyPathValid {
		return certPath, keyPath, nil
	}

	synthesizer := Synthesizer{}
	selfSignedCertPem, selfSignedKeyPem, err := synthesizer.SelfSignedCertificatePairPemFactory(nil, nil)
	if err != nil {
		return certPath, keyPath, err
	}

	fileClerk := FileClerk{}
	pkiDir := "pki"
	if os.Getenv(ReadThroughPkiDirEnvVarName) != "" {
		pkiDir = os.Getenv(ReadThroughPkiDirEnvVarName)
	}
	err = fileClerk.CreateDir(pkiDir)
	if err != nil {
		return certPath, keyPath, err
	}

	rawCertPath, err = filepath.Abs(pkiDir + "/cert.pem")
	if err != nil {
		return certPath, keyPath, errors.New("RetrieveSelfSignedCertAbsolutePathFailed")
	}
	certPath, err = tkValueObject.NewUnixAbsoluteFilePath(rawCertPath, allowUnsafeFilePathChars)
	if err != nil {
		return certPath, keyPath, errors.New("SelfSignedCertPathInvalid")
	}
	err = fileClerk.UpdateFileContent(certPath.String(), selfSignedCertPem, true)
	if err != nil {
		return certPath, keyPath, errors.New("SelfSignedCertContentUpdateFailed")
	}

	rawKeyPath, err = filepath.Abs(pkiDir + "/key.pem")
	if err != nil {
		return certPath, keyPath, errors.New("RetrieveSelfSignedKeyAbsolutePathFailed")
	}
	keyPath, err = tkValueObject.NewUnixAbsoluteFilePath(rawKeyPath, allowUnsafeFilePathChars)
	if err != nil {
		return certPath, keyPath, errors.New("SelfSignedKeyPathInvalid")
	}
	err = fileClerk.UpdateFileContent(keyPath.String(), selfSignedKeyPem, true)
	if err != nil {
		return certPath, keyPath, errors.New("SelfSignedKeyContentUpdateFailed")
	}

	return certPath, keyPath, nil
}
