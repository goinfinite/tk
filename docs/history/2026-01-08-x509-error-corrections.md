# X509 Certificate Implementation - Error Corrections

**Date**: 2026-01-08
**Context**: Post-implementation review after creating 24 X509 value objects and entity constructors

## Session 1

1. **Missing Plan File**: Implementation started without saving plan to docs/history/
2. **Error Naming**: Used "FailedTo" prefix → Changed to suffix pattern ("ParseCertificateFailed")
3. **Generic Variable Names**: `cert` → `stdlibCert`, `block` → `pemBlock`, `notBefore` → `validityNotBeforeUnix`
4. **Missing Ptr Suffix**: `issuerCommonName` → `issuerCommonNamePtr`
5. **Mysterious Inline Commands**: Extracted `hex.EncodeToString(sha256Hash[:])` to `sha256FingerprintHex` variable
6. **Non-Self-Explanatory Conditions**: `stdlibCert.MaxPathLen >= 0` → `maxPathLengthIsSet := ...`
7. **Locality of Behavior Violation**: Package-level functions → Moved to value object constructors (`NewX509DistinguishedNameFromPkixName`)

## Session 2

1. **Error Corrections Not Archived**: Created this document
2. **Type Conversion as Inline Cast**: Added `Bytes() []byte` method to X509EnvelopedCertificate
3. **Variable Name Ambiguity with Package**: `x509Cert` → `x509CertEntity` (package name collision)
4. **Variable Names Describing Content**: Need to focus on intention/purpose rather than what the value is
5. **Incomplete Implementation**: Certificate policies left empty → Implemented parsing from stdlib
6. **Missing Final Review Step**: Added "After Completion" section to DEV-IMPLEMENTATION.md

## Session 3

1. **Empty Slice vs Nil**: `[]X509PolicyQualifier{}` → `nil` (unavailable optional data should be nil)
2. **Single-Use Intermediate Variable**: Removed unnecessary intermediate variable for single-use cast
3. **Technical Abbreviations**: `certificateIsCA` → `certIsAuthority` (CA is domain-specific)
4. **Missing Debug Logging**: Added `slog.Debug` for all skipped items in loops
5. **Inconsistent Field Naming**: `IsCA` → `IsAuthority` across all files
6. **Function Parameter Formatting**: Split different types to separate lines in declarations
7. **Unnecessary Line Breaks**: Consolidated function instantiations to reduce unnecessary breaks
8. **Inefficient String Building**: Replaced slice append + strings.Join with strings.Builder

## Session 4

1. **Unvalidated Data Naming**: `organizationalUnitStr` → `rawOrganizationalUnit` (unvalidated = "raw" prefix)
2. **Generic Variable in Return**: `cert` → `envelopedCert` in named return
3. **Missing slog.Debug**: Added debug logging in x509DistinguishedName.go, x509ExtendedKeyUsage.go (2 locations), x509KeyUsage.go, x509Certificate.go
4. **"Str" Suffix Misuse**: `subjectCommonNameStr` → `rawSubjectCommonName` ("Str" only for .String() results, not raw data)

**Note**: Initial review missed 6 violations - need systematic checking of ALL files/loops.

## Session 5

1. **Missing Accent Normalization**: Created `util/stringNormalizer.go` with StripAccents() using golang.org/x/text NFD/NFC transformation
2. **Weak x509PolicyName Validation**: Added regex `^[a-zA-Z0-9 .\-_()]{1,128}$` for character whitelisting
3. **Weak Wildcard Validation**: Added format check with regex `^\*\.[a-zA-Z0-9.\-_]+$` beyond just counting wildcards
4. **Unnecessary Line Breaks in Tests**: 90+ instances across 23 test files - consolidated multi-line t.Errorf/t.Fatalf to single lines
5. **Forgot to Document** (2nd occurrence)

## Session 6

1. **Missing Test Coverage**: x509SignatureValue allows `\r\n` but no tests verified it - added 3 test cases
2. **CRITICAL else Violation**: x509SubjectName.go used else block → Replaced with early return
3. **Incomplete Test Coverage**: StripAccents used in 4 VOs but only Locality tested → Added tests to Organization, OrganizationalUnit, StateOrProvince
4. **Forgot to Document** (3rd occurrence)

## Session 7

1. **Not Using Table-Driven Tests**: Entity test used t.Run() subtests → Refactored to testCaseStructs pattern
2. **Hardcoded Test Values**: `!= 3`, `!= "RSA"` → Extracted to `expectedVersion`, `expectedPublicKeyAlgorithm`
3. **Generic Test Variable**: `x509Cert` → `x509CertEntity`
4. **Missing continue**: Added continue after error validation to skip success assertions
5. **Forgot to Document** (4th occurrence)

## Session 8

1. **Security Vulnerability**: strings.Contains allowed injection attacks → Replaced with regex `^-----BEGIN CERTIFICATE-----[\s\S]+-----END CERTIFICATE-----$`
   - Added 4 security test cases (XSS, command injection, data before/after PEM)
2. **Incorrect Regex**: Optional groups allowed mismatched BEGIN/END tags → Split into 5 explicit patterns + strings.Count validation
   - Go's RE2 doesn't support backreferences or negative lookahead
3. **Missing Security Tests**: Added attack scenario test coverage

## Summary Statistics

**Total Violations**: 48 across 8 sessions
- Session 1: 8 (naming, organization)
- Session 2: 7 (naming, implementation)
- Session 3: 9 (initialization, naming, logging, formatting)
- Session 4: 7 (naming semantics, logging)
- Session 5: 5 (security, UX, formatting)
- Session 6: 4 (test coverage, else, documentation)
- Session 7: 5 (test pattern, naming, documentation)
- Session 8: 3 (security, regex)

## Key Recurring Issues

1. **Generic/Ambiguous Names** (5 sessions): Most persistent technical error
2. **Documentation Forgetting** (4 sessions, 57% failure): Systematic issue
3. **Missing slog.Debug** (2 sessions): Debug logging for skipped items
4. **Test Coverage Gaps** (2 sessions): Incomplete testing
5. **Security Validation** (2 sessions): Injection prevention

## Rules Added

1. Use regex with anchors (^, $) for format validation; strings.Contains only checks presence
2. When regex has optional groups that must match, use separate explicit patterns
3. Go's RE2 doesn't support negative lookahead - use strings.Count or explicit validation
4. Security-sensitive validation needs test cases for attack scenarios
5. Variable naming semantics: `raw*` = unvalidated, `*Str` = .String() result, `stdlib*` = from Go stdlib
6. Always use table-driven test pattern
7. Extract expected test values to named variables
8. Use `continue` in table-driven tests to skip success assertions after error cases
