# X509 Certificate Value Objects Implementation Plan

## Overview

Complete all TODOs in `src/domain/entity/x509Certificate.go` by creating 24 missing value objects and their unit tests, plus implementing constructors and methods for the entity types.

## Implementation Approach

- **Validation**: Practical/lenient validation that works with real-world certificates
- **Certificate Parsing**: Full implementation using Go's crypto/x509 standard library
- **Composite Value Objects**: X509DistinguishedName and X509CertificatePolicy will be moved to valueObject package as struct-based value objects (per user request, breaking the typical simple-type pattern)

## Current State

- File: `/var/home/ntorga/infinite/projects/tk/src/domain/entity/x509Certificate.go`
- 24 X509-related types are referenced but not implemented
- All value objects need to be created from scratch
- No existing X509 value objects in the codebase

## Value Objects to Create

### 1. String-Based Value Objects (Simple Validation)

#### X509Organization

- **Type**: `string`
- **Validation**: Non-empty, 1-255 chars, alphanumeric + spaces + common punctuation
- **Pattern**: Similar to organization names in certificates
- **File**: `x509Organization.go`

#### X509OrganizationalUnit

- **Type**: `string`
- **Validation**: Non-empty, 1-255 chars, alphanumeric + spaces + common punctuation
- **Pattern**: Similar to X509Organization
- **File**: `x509OrganizationalUnit.go`

#### X509Locality

- **Type**: `string`
- **Validation**: Non-empty, 1-128 chars, letters + spaces + hyphens
- **Pattern**: City/town name validation
- **File**: `x509Locality.go`

#### X509StateOrProvince

- **Type**: `string`
- **Validation**: Non-empty, 1-128 chars, letters + spaces + hyphens
- **Pattern**: Similar to X509Locality
- **File**: `x509StateOrProvince.go`

#### X509SubjectName

- **Type**: `string`
- **Validation**: Hostname or common name format (1-253 chars)
- **Pattern**: Can be FQDN, wildcard domain (\*.example.com), or common name
- **File**: `x509SubjectName.go`

#### X509PolicyName

- **Type**: `string`
- **Validation**: Non-empty, 1-128 chars
- **File**: `x509PolicyName.go`

### 2. String-Based with Hex Validation

#### X509SerialNumber

- **Type**: `string`
- **Validation**: Hex string, 1-40 chars (up to 160 bits)
- **Regex**: `^[0-9A-Fa-f]{1,40}$`
- **Pattern**: Similar to `hash.go` but with specific X509 serial number constraints
- **File**: `x509SerialNumber.go`

#### X509Fingerprint

- **Type**: `string`
- **Validation**: Hex string, exactly 40 chars (SHA1) or 64 chars (SHA256)
- **Regex**: `^[0-9A-Fa-f]{40}$|^[0-9A-Fa-f]{64}$`
- **Pattern**: Similar to `hash.go`
- **File**: `x509Fingerprint.go`

#### X509KeyIdentifier

- **Type**: `string`
- **Validation**: Hex string, typically 40 chars (20 bytes)
- **Regex**: `^[0-9A-Fa-f]{40}$`
- **File**: `x509KeyIdentifier.go`

#### X509SignatureValue

- **Type**: `string`
- **Validation**: Hex string or base64, 64-1024 chars
- **File**: `x509SignatureValue.go`

### 3. String-Based with Special Format

#### X509PolicyOID

- **Type**: `string`
- **Validation**: Dotted notation (e.g., "1.2.840.113549.1.1.11")
- **Regex**: `^[0-9]+(\.[0-9]+)*$`
- **Min components**: 2
- **File**: `x509PolicyOid.go`

#### X509EnvelopedCertificate

- **Type**: `string`
- **Validation**: PEM format certificate (starts with `-----BEGIN CERTIFICATE-----`)
- **Pattern**: Multi-line string validation
- **File**: `x509EnvelopedCertificate.go`

#### X509PublicKeyValue

- **Type**: `string`
- **Validation**: Base64 encoded public key
- **Length**: Variable, typically 100-800+ chars
- **File**: `x509PublicKeyValue.go`

### 4. Enum-Based Value Objects

#### X509PublicKeyAlgorithm

- **Type**: `string` enum
- **Values**:
  - `RSA`
  - `ECDSA`
  - `Ed25519`
  - `DSA`
- **Pattern**: Similar to `dnsRecordType.go`
- **File**: `x509PublicKeyAlgorithm.go`

#### X509SignatureAlgorithm

- **Type**: `string` enum
- **Values**:
  - `SHA256WithRSA`
  - `SHA384WithRSA`
  - `SHA512WithRSA`
  - `ECDSAWithSHA256`
  - `ECDSAWithSHA384`
  - `ECDSAWithSHA512`
  - `Ed25519`
- **Pattern**: Similar to `dnsRecordType.go`
- **File**: `x509SignatureAlgorithm.go`

#### X509KeyUsage

- **Type**: `string` enum
- **Values**:
  - `digitalSignature`
  - `contentCommitment` (formerly nonRepudiation)
  - `keyEncipherment`
  - `dataEncipherment`
  - `keyAgreement`
  - `keyCertSign`
  - `cRLSign`
  - `encipherOnly`
  - `decipherOnly`
- **Pattern**: Similar to `dnsRecordType.go`
- **File**: `x509KeyUsage.go`

#### X509ExtendedKeyUsage

- **Type**: `string` enum
- **Values**:
  - `serverAuth`
  - `clientAuth`
  - `codeSigning`
  - `emailProtection`
  - `timeStamping`
  - `ocspSigning`
- **Pattern**: Similar to `dnsRecordType.go`
- **File**: `x509ExtendedKeyUsage.go`

#### X509PolicyQualifier

- **Type**: `string` enum
- **Values**:
  - `cps` (Certification Practice Statement)
  - `userNotice`
- **File**: `x509PolicyQualifier.go`

### 5. Numeric Value Objects

#### X509VersionNumber

- **Type**: `uint8`
- **Validation**: Must be 1, 2, or 3
- **Pattern**: Similar to `networkPort.go` but with enum validation
- **File**: `x509VersionNumber.go`

#### X509PublicKeySize

- **Type**: `uint16`
- **Validation**: Common key sizes (1024, 2048, 3072, 4096, 8192, 256, 384, 521)
- **Pattern**: Similar to `networkPort.go` with switch validation
- **File**: `x509PublicKeySize.go`

### 6. Struct-Based Value Objects

#### X509BasicConstraints

- **Type**: struct with:
  - `IsCA` (bool)
  - `MaxPathLength` (\*int - pointer for optional)
- **Validation**: MaxPathLength >= 0 when present
- **Note**: This is a composite value object (breaking typical simple-type pattern per user request)
- **File**: `x509BasicConstraints.go`

#### X509DistinguishedName

- **Type**: struct with:
  - `Organization` (\*X509Organization)
  - `OrganizationalUnit` ([]X509OrganizationalUnit)
  - `Locality` (\*X509Locality)
  - `StateOrProvince` (\*X509StateOrProvince)
  - `Country` (\*CountryCode)
- **Constructor**: `NewX509DistinguishedName`
- **String() method**: Returns standard format `"O=Organization, OU=Unit, L=Locality, ST=State, C=Country"`
- **Note**: Moving from entity to valueObject per user request
- **File**: `x509DistinguishedName.go`

#### X509CertificatePolicy

- **Type**: struct with:
  - `PolicyIdentifier` (X509PolicyOID)
  - `PolicyName` (\*X509PolicyName)
  - `PolicyQualifiers` ([]X509PolicyQualifier)
- **Constructor**: `NewX509CertificatePolicy`
- **Note**: Moving from entity to valueObject per user request
- **File**: `x509CertificatePolicy.go`

## Entity Updates

### File: `src/domain/entity/x509Certificate.go`

#### 1. Update imports and references

- Change X509DistinguishedName references from `entity.X509DistinguishedName` to `tkValueObject.X509DistinguishedName`
- Change X509CertificatePolicy references similarly
- Keep entity structure otherwise unchanged

#### 2. X509Certificate Constructor

- Add `NewX509Certificate` constructor
- Parameters: All required fields (non-pointer)
- Return: `X509Certificate` entity

#### 3. X509CertificateFromEnvelopedCertificate Constructor

- Add `NewX509CertificateFromEnvelopedCertificate` constructor
- Parameter: `X509EnvelopedCertificate`
- **Full implementation**: Parse PEM certificate using `crypto/x509` and `encoding/pem` packages
- Extract all certificate fields and create value objects for each
- Return: `(X509Certificate, error)`
- Handle parsing errors gracefully

## Implementation Steps

### Phase 1: Simple String Value Objects (6 files)

1. X509Organization + test
2. X509OrganizationalUnit + test
3. X509Locality + test
4. X509StateOrProvince + test
5. X509SubjectName + test
6. X509PolicyName + test

### Phase 2: Hex String Value Objects (4 files)

1. X509SerialNumber + test
2. X509Fingerprint + test
3. X509KeyIdentifier + test
4. X509SignatureValue + test

### Phase 3: Special Format String Value Objects (3 files)

1. X509PolicyOID + test
2. X509EnvelopedCertificate + test
3. X509PublicKeyValue + test

### Phase 4: Enum Value Objects (5 files)

1. X509PublicKeyAlgorithm + test
2. X509SignatureAlgorithm + test
3. X509KeyUsage + test
4. X509ExtendedKeyUsage + test
5. X509PolicyQualifier + test

### Phase 5: Numeric Value Objects (2 files)

1. X509VersionNumber + test
2. X509PublicKeySize + test

### Phase 6: Complex/Composite Value Objects (3 files)

1. X509BasicConstraints + test
2. X509DistinguishedName + test (moved from entity)
3. X509CertificatePolicy + test (moved from entity)

### Phase 7: Entity Updates

1. Update x509Certificate.go entity imports and type references
2. Add NewX509Certificate constructor
3. Add NewX509CertificateFromEnvelopedCertificate constructor with full PEM parsing using crypto/x509

## File Locations

- Value objects: `/var/home/ntorga/infinite/projects/tk/src/domain/valueObject/`
- Tests: Same directory as implementation files
- Entity: `/var/home/ntorga/infinite/projects/tk/src/domain/entity/x509Certificate.go`

## Testing Strategy

- Follow table-driven test pattern from existing value objects
- Test constructor with valid and invalid inputs
- Test String() method
- Test type conversion methods where applicable
- Include edge cases and security inputs (XSS, command injection attempts)

## Summary of Deliverables

- **24 value object files** (implementation)
- **24 test files** (unit tests)
- **1 entity file update** (x509Certificate.go with constructors)
- **Total**: 49 files created/modified

## Validation After Implementation

1. Run all tests: `go test ./src/domain/valueObject/x509*_test.go -v`
2. Verify no compilation errors in x509Certificate.go entity file: `go build ./src/domain/entity/x509Certificate.go`
3. Run `go mod tidy` to ensure dependencies are correct
4. Verify all TODOs in x509Certificate.go are resolved
5. Test NewX509CertificateFromEnvelopedCertificate with a real PEM certificate

## Critical Files

- `/var/home/ntorga/infinite/projects/tk/src/domain/entity/x509Certificate.go` (entity with TODOs)
- `/var/home/ntorga/infinite/projects/tk/src/domain/valueObject/*.go` (21+ new value objects)
- `/var/home/ntorga/infinite/projects/tk/docs/DEV-IMPLEMENTATION.md` (coding standards)
- `/var/home/ntorga/infinite/projects/tk/src/domain/valueObject/.context.md` (layer rules)

## Notes

- All value objects must follow the constructor pattern: `NewX509TypeName(value any) (X509TypeName, error)`
- Error messages: `"X509TypeNameMustBe[Type]"` and `"InvalidX509TypeName"`
- Use tkVoUtil for type conversions
- All files use package `tkValueObject`
- Import alias: `tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"`
