# Domain

Business logic layer of Infinite Toolkit _(TK)_. It holds validated value objects, entities, DTOs, repository interfaces, and use cases. The domain layer never imports infra or presentation. Part of [Infinite Toolkit](../../README.md).

## Value Objects

The library offers a diverse range of value objects (VO) to represent domain entities. Each VO is designed to guarantee type safety and provide validation. Examples include Email, Password, URL, IPAddress, UnixFilePath, HttpMethod, CountryCode, CurrencyCode, SystemResourceIdentifier, and many more. These components are thoroughly tested, ensuring 100% coverage.

## Value Object Utilities

- **InterfaceTo**: Safely convert interface{} to primitive types (bool, string, int, float, etc.) using reflection, handling various input formats with error checking.

  ```go
  boolValue, boolConversionErr := tkVoUtil.InterfaceToBool("true")
  stringValue, stringConversionErr := tkVoUtil.InterfaceToString(42)
  intValue, intConversionErr := tkVoUtil.InterfaceToInt("123")
  int8Value, int8ConversionErr := tkVoUtil.InterfaceToInt8("127")
  int16Value, int16ConversionErr := tkVoUtil.InterfaceToInt16("32767")
  int32Value, int32ConversionErr := tkVoUtil.InterfaceToInt32("2147483647")
  int64Value, int64ConversionErr := tkVoUtil.InterfaceToInt64(3.14159)
  uintValue, uintConversionErr := tkVoUtil.InterfaceToUint("4294967295")
  uint8Value, uint8ConversionErr := tkVoUtil.InterfaceToUint8("255")
  uint16Value, uint16ConversionErr := tkVoUtil.InterfaceToUint16("65535")
  uint32Value, uint32ConversionErr := tkVoUtil.InterfaceToUint32("4294967295")
  uint64Value, uint64ConversionErr := tkVoUtil.InterfaceToUint64("18446744073709551615")
  float32Value, float32ConversionErr := tkVoUtil.InterfaceToFloat32("-987.654")
  float64Value, float64ConversionErr := tkVoUtil.InterfaceToFloat64("-123.456")
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
      createDto dto.CreateSessionToken,
  ) (attemptsCount uint, err error) {
      failedAttemptsIntervalStartsAt := tkValueObject.NewUnixTimeBeforeNow(
          CreateSessionTokenFailedLoginAttemptsInterval,
      )
      readResponseDto, err := tkUseCase.ReadActivityRecords(
          activityRecordQueryRepo, tkDto.ReadActivityRecordsRequest{
              Pagination:        tkUseCase.ActivityRecordsDefaultPagination,
              RecordLevel:       &tkValueObject.ActivityRecordLevelSecurity,
              RecordCode:        &CreateSessionTokenActivityRecordCodeFailed,
              OperatorIpAddress: &createDto.OperatorIpAddress,
              CreatedAfterAt:    &failedAttemptsIntervalStartsAt,
          })
      if err != nil {
          return attemptsCount, err
      }

      return uint(len(readResponseDto.ActivityRecords)), nil
  }
  ```
