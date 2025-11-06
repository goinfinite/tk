# Changelog

```log
0.1.1 - 2025/11/06
feat: add request input reader
feat: add trusted ip reader
feat: add pagination parser
feat: add time params parser
feat: add response wrappers
fix: turn last seen id vo regex stricter

0.1.0 - 2025/11/05
feat: add SelfSignedCertificatePairFactory to Synthesizer
feat: add CertPairFilePathsReader to ReadThrough
feat: add envsInspector presentation helper
fix: decompress using source dir as working dir
fix: keep only utf8 chars on StripUnsafe

0.0.9 - 2025/11/03
feat: split unix file path into relative and absolute vos
feat: add panic handler middleware
feat: add log handler middleware
chore: add echo as dependency
chore: add zerolog as dependency

0.0.8 - 2025/10/31
feat: import, refactor and create unit tests for common vos from OS/Ez/Bz projects
fix: move regex must compile to pkg level

0.0.7 - 2025/06/17
feat: add FileClerk
feat: add CompressionFormat vo
fix: add stdout and stderr file handlers to shell

0.0.6 - 2025/06/11
feat: add Shell
feat: add ReadServerPublic/PrivateIpAddresses
feat: add IsBetween() for UnixTime

0.0.5 - 2025/06/02
chore: remove RequestInputParser
fix: StringSliceValueObjectParser nil and empty string check

0.0.4 - 2025/06/01
feat: add UnixTime vo
feat: add RequiredParamsInspector

0.0.3 - 2025/05/31
feat: add RequestInputParser

0.0.2 - 2025/05/16
feat: add deserializer

0.0.1 - 2025/05/11
feat: initial release
```
