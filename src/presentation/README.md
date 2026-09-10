# Presentation

Presentation layer of Infinite Toolkit _(TK)_. It parses untrusted input, wraps responses, and provides Echo middleware. It depends on infra and domain. Part of [Infinite Toolkit](../../README.md).

## Middleware

- **PanicHandler**: Handle panics in HTTP requests, log filtered stack traces (excluding domain layers), and respond with error messages; uses `TRUSTED_IPS` and `TRUSTED_CIDRS` env vars to mask sensitive information from untrusted operators. It can be used with CLI applications as well.

  ```go
  // ForApiInitialization
  echoInstance.Use(ApiPanicHandler)

  // ForCliInitialization
  defer CliPanicHandler()
  ```

- **LogHandler**: Configure logging levels via `LOG_LEVEL` environment variable and initialize structured logging with slog and Zerolog.

## Helpers

- **RequiredParamsInspector**: Check that every required parameter is present in a request input map before further parsing.

  ```go
  paramsReceived := map[string]any{"name": "John", "age": 30}
  paramsRequired := []string{"name", "email"}
  requiredParamsValidationErr := RequiredParamsInspector(paramsReceived, paramsRequired)
  ```

- **ApiRequestInputReader**: Read and parse JSON, form data, or multipart files from Echo HTTP requests into structured data.

  ```go
  inputReader := ApiRequestInputReader{}
  requestData, requestParsingErr := inputReader.Reader(echoContext)
  ```

- **RequesterIpExtractor**: Headers-first IP extraction with a right-to-left trust chain walk. Reads `IP_EXTRACT_HEADER` env var as a comma-separated ordered chain (default: `X-Forwarded-For,X-Real-IP`). Each header is walked right-to-left; the first entry that is not a trusted IP (local, private, link-local, or in `TRUSTED_IPS`/`TRUSTED_CIDRS`) is returned. The chain supports the special keywords `Direct` and `RemoteAddr` to force extraction from `http.Request.RemoteAddr`. Falls back implicitly to `RemoteAddr` when all headers are exhausted.

  > **Security:** Right-to-left traversal means values appended by untrusted clients are skipped automatically. The actual risk is `TRUSTED_CIDRS` configured too broadly, which would cause the extractor to skip a proxy range that includes untrusted clients.

  ```go
  extractor := NewRequesterIpExtractor()

  ipAddress, err := extractor.Execute(httpRequest)
  ```

- **EnvsInspector**: Inspect and validate environment variables from .env files loaded via `ENV_FILE_PATH` env var, supporting required and auto-fillable variables.

  ```go
  optionalEnvFilePath := "/path/to/.env"
  requiredEnvVarNames := []string{"TRAIL_DATABASE_FILE_PATH", "SESSION_TOKEN_SECRET"}
  autoFillableEnvVars := []string{"SESSION_TOKEN_SECRET"}
  envsInspector := NewEnvsInspector(
    optionalEnvFilePath, requiredEnvVarNames, autoFillableEnvVars,
  )
  envsValidationErr := envsInspector.Inspect()
  ```

- **PaginationParser**: Parse pagination parameters like pageNumber, itemsPerPage, lastSeenId, sortBy, and sortDirection from HTTP requests.

  ```go
  defaultPagination := tkDto.Pagination{PageNumber: 0, ItemsPerPage: 10}
  untrustedInput := map[string]any{"pageNumber": 1, "itemsPerPage": 20}
  parsedPagination, paginationParsingErr := PaginationParser(
    defaultPagination, untrustedInput,
  )
  ```

- **StringSliceVoParser**: Convert comma-separated, semicolon-separated, or array strings into value object slices.

  ```go
  rawInput := "tag1,tag2;tag3"
  parsedTags := StringSliceValueObjectParser(rawInput, tkValueObject.NewTag)
  ```

- **TimeParamsParser**: Parse date ranges, timestamps, and relative times from request parameters.

  ```go
  timeParamNames := []string{"createdAt", "updatedAt"}
  untrustedInput := map[string]any{"createdAt": 1609459200}
  parsedTimeParams := TimeParamsParser(timeParamNames, untrustedInput)
  fmt.Println(parsedTimeParams["createdAt"])
  ```

- **ResponseWrapper**: A wrapper struct for liaison responses and when needed, used to emit API and CLI responses.

  ```go
  apiResponse := NewApiResponseWrapper(201, accountEntity, "AccountCreatedSuccessfully")

  liaisonResponse := NewLiaisonResponse(
    LiaisonResponseStatusCreated, accountEntity, "AccountCreatedSuccessfully",
  )

  liaisonResponseNoMessage := NewLiaisonResponseNoMessage(
    LiaisonResponseStatusSuccess, err.Error(),
  )

  liaisonApiEmissionErr := LiaisonApiResponseEmitter(echoContext, liaisonResponse)

  LiaisonCliResponseRenderer(liaisonResponse)
  ```
