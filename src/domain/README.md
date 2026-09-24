# Domain

Business logic layer of Infinite Toolkit _(TK)_. It holds validated value objects, entities, DTOs, repository interfaces, and use cases. The domain layer never imports infra or presentation. Part of [Infinite Toolkit](../../README.md).

## Value Objects

The library offers a diverse range of value objects (VO) to represent domain entities. Each VO is designed to guarantee type safety and provide validation. Examples include MailAddress, Password, Url, IpAddress, UnixAbsoluteFilePath, HttpMethod, CountryCode, CurrencyCode, SystemResourceIdentifier, CatalogItemName, ScheduledTaskName, CronSchedule, and many more. Table-driven tests cover the validation rules.

## Value Object Utilities

- **InterfaceTo**: Safely convert any value to primitive types (bool, string, int, float, etc.) using reflection, handling various input formats with error checking.

  ```go
  boolValue, boolConversionErr := tkVoUtil.InterfaceToBool("true")
  stringValue, stringConversionErr := tkVoUtil.InterfaceToString(42)
  intValue, intConversionErr := tkVoUtil.InterfaceToInt("123")
  int8Value, int8ConversionErr := tkVoUtil.InterfaceToInt8("127")
  int16Value, int16ConversionErr := tkVoUtil.InterfaceToInt16("32767")
  int32Value, int32ConversionErr := tkVoUtil.InterfaceToInt32("2147483647")
  int64Value, int64ConversionErr := tkVoUtil.InterfaceToInt64("123")
  uintValue, uintConversionErr := tkVoUtil.InterfaceToUint("4294967295")
  uint8Value, uint8ConversionErr := tkVoUtil.InterfaceToUint8("255")
  uint16Value, uint16ConversionErr := tkVoUtil.InterfaceToUint16("65535")
  uint32Value, uint32ConversionErr := tkVoUtil.InterfaceToUint32("4294967295")
  uint64Value, uint64ConversionErr := tkVoUtil.InterfaceToUint64("18446744073709551615")
  float32Value, float32ConversionErr := tkVoUtil.InterfaceToFloat32("-987.654")
  float64Value, float64ConversionErr := tkVoUtil.InterfaceToFloat64("-123.456")
  ```

- **IsFractionalFloat**: Report whether a float input has a fractional part, so integer-domain constructors can reject it before conversion.

  ```go
  isFractional := tkVoUtil.IsFractionalFloat(42.5)
  ```

- **TruncateFloat**: Truncate a float input to its whole part; non-float input passes through unchanged.

  ```go
  truncated := tkVoUtil.TruncateFloat(42.9)
  ```

- **SafeTruncateString**: Truncate a string to a byte limit without splitting a UTF-8 rune.

  ```go
  truncated := tkVoUtil.SafeTruncateString("command output", 4096)
  ```

- **NamedGroupsExtractor**: Extract named capture groups from a regex match into a map keyed by group name.

  ```go
  namedGroups := tkVoUtil.NamedGroupsExtractor(urlRegex, "https://example.com")
  ```

- **StripAccents**: Remove diacritical marks from a string and trim it.

  ```go
  normalized, stripErr := tkVoUtil.StripAccents("Café")
  ```

- **StripHexSeparators**: Remove colons and spaces from a hexadecimal string.

  ```go
  normalized := tkVoUtil.StripHexSeparators("AA:BB:CC")
  ```

- **IsAnyError**: Report whether an error matches any of the target errors through `errors.Is`.

  ```go
  isTimeout := tkVoUtil.IsAnyError(err, context.DeadlineExceeded, context.Canceled)
  ```

## DTOs

- **Pagination**: General pagination DTO with page number, items per page, last seen ID, sort by, and sort direction.

## Activity Record Management

Infinite Toolkit _(TK)_ provides a comprehensive activity record management system for auditing and logging user actions, following Clean Architecture principles. The GORM implementations of the repository interfaces live in `src/infra/activityRecord`.

### Entity

- **ActivityRecord**: Represents an activity record with record ID, level, code, affected resources, details, operator account ID, IP address, and creation time.

### Use Cases

- **CreateActivityRecord**: Persists an activity record as a non-blocking side effect, logging errors without failing the primary operation.

  ```go
  recordCode, _ := tkValueObject.NewActivityRecordCode("CreateAccount")
  affectedResources := []tkValueObject.SystemResourceIdentifier{
      tkValueObject.NewSriAccount(2),
  }
  operatorSri := tkValueObject.NewSriAccount(1)
  operatorIpAddress, _ := tkValueObject.NewIpAddress("1.1.1.1")

  createDto := tkDto.CreateActivityRecord{
      RecordLevel:       tkValueObject.ActivityRecordLevelSecurity,
      RecordCode:        recordCode,
      AffectedResources: affectedResources,
      RecordDetails:     map[string]any{"username": "abc123"},
      OperatorSri:       &operatorSri,
      OperatorIpAddress: &operatorIpAddress,
  }

  tkUseCase.CreateActivityRecord(activityRecordCmdRepo, createDto)
  ```

- **DeleteActivityRecord**: Deletes an activity record by ID.

  ```go
  recordId, _ := tkValueObject.NewActivityRecordId(123)
  deleteDto := tkDto.DeleteActivityRecord{RecordId: &recordId}
  deleteErr := tkUseCase.DeleteActivityRecord(activityRecordCmdRepo, deleteDto)
  ```

- **ReadActivityRecords**: Retrieves activity records with pagination and filtering options.

  ```go
  requestDto := tkDto.ReadActivityRecordsRequest{
      Pagination: tkDto.Pagination{
          PageNumber:   0,
          ItemsPerPage: 20,
      },
  }

  responseDto, readErr := tkUseCase.ReadActivityRecords(
    activityRecordQueryRepo, requestDto,
  )
  activityRecords := responseDto.ActivityRecords
  ```

### DTOs

- **CreateActivityRecord**: Data transfer object for creating activity records.
- **DeleteActivityRecord**: Data transfer object for deleting activity records.
- **ReadActivityRecords**: Data transfer object for reading activity records with pagination.

### Repositories

- **ActivityRecordCmdRepo**: Interface for command operations (create, delete) on activity records.
- **ActivityRecordQueryRepo**: Interface for query operations (read) on activity records.

### Usage Examples

- **Counting Failed Login Attempts**: Query activity records to count failed login attempts for security monitoring.

  ```go
  func readFailedLoginAttemptsCount(
      activityRecordQueryRepo tkRepository.ActivityRecordQueryRepo,
      operatorIpAddress tkValueObject.IpAddress,
  ) (attemptsCount uint, err error) {
      recordLevel := tkValueObject.ActivityRecordLevelSecurity
      recordCode, _ := tkValueObject.NewActivityRecordCode(
          "CreateSessionTokenFailed",
      )
      failedAttemptsIntervalStartsAt := tkValueObject.NewUnixTimeBeforeNow(
          24 * time.Hour,
      )

      readResponseDto, err := tkUseCase.ReadActivityRecords(
          activityRecordQueryRepo, tkDto.ReadActivityRecordsRequest{
              Pagination:        tkUseCase.ActivityRecordsDefaultPagination,
              RecordLevel:       &recordLevel,
              RecordCode:        &recordCode,
              OperatorIpAddress: &operatorIpAddress,
              CreatedAfterAt:    &failedAttemptsIntervalStartsAt,
          })
      if err != nil {
          return attemptsCount, err
      }

      return uint(len(readResponseDto.ActivityRecords)), nil
  }
  ```
