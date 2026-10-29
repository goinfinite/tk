# Presentation

Presentation layer of Infinite Toolkit _(TK)_. It parses untrusted input, wraps responses, and provides Echo middleware. It depends on infra and domain. Part of [Infinite Toolkit](../../README.md).

## Middleware

- **PanicHandler**: Handle panics in HTTP requests, log filtered stack traces (excluding domain layers), and respond with error messages; uses `TRUSTED_IPS` and `TRUSTED_CIDRS` env vars to mask sensitive information from untrusted operators. It can be used with CLI applications as well. The trust check uses the extracted requester IP. When the whole chain is trusted, that IP is the original requester as recorded by the first proxy. A client on the trusted network can claim another trusted IP by sending its own `X-Forwarded-For`. Overwrite the header at the proxy to prevent this.

  ```go
  // ForApiInitialization
  echoInstance.Use(ApiPanicHandler)

  // ForCliInitialization
  defer CliPanicHandler()
  ```

- **LogHandler**: Configure logging levels via `LOG_LEVEL` environment variable and initialize structured logging with slog and Zerolog.

- **ErrorStatusHandler**: Map a handler error to an HTTP status code and emit the standard API response envelope. The constructor takes `map[error]tkValueObject.HttpStatusCode`. A nil error and a direct `*echo.HTTPError` pass through unchanged. A committed response is not overwritten; the error is logged and returned. A returned error matches through `errors.Is`, so a wrapped sentinel error keeps its mapping. A match responds with the mapped status and the matched sentinel's message, so a wrapper's text does not reach the client. A mapped status outside 400-599 fails closed to 500 with a generic message. An unmapped error or a mapped 5xx is logged at error level. An ambiguous match or an invalid mapped status is logged at warn level. Log entries include the request method, path, route, request ID, and operator identity when available. An unmapped or ambiguous error responds with 500 and a generic message.

  ```go
  errorStatusHandler := NewErrorStatusHandler(map[error]tkValueObject.HttpStatusCode{
      errRecordNotFound: tkValueObject.HttpStatusCodeNotFound,
  })
  echoInstance.Use(errorStatusHandler.ApiHandler)
  ```

## Helpers

- **RequiredParamsInspector**: Check that every required parameter is present in a request input map before further parsing.

  ```go
  paramsReceived := map[string]any{"name": "John", "age": 30}
  paramsRequired := []string{"name", "email"}
  requiredParamsValidationErr := RequiredParamsInspector(paramsReceived, paramsRequired)
  ```

- **ApiRequestInputReader**: Read and parse JSON, form data, or multipart files from Echo HTTP requests into structured data. Uploaded files are grouped under the `files` key by form field: `requestBody["files"]` is a `map[string][]*multipart.FileHeader`, one key per field, each slice in upload order. A request with no files carries no `files` key.

  ```go
  inputReader := ApiRequestInputReader{}
  requestData, requestParsingErr := inputReader.Reader(echoContext)
  ```

- **RequesterIpExtractor**: Requester IP extraction with a peer trust gate and a right-to-left trust chain walk. Reads `IP_EXTRACT_HEADER` env var as a comma-separated ordered chain (default: `X-Forwarded-For,X-Real-IP`). When `RemoteAddr` is not a trusted IP (local, private, link-local, or in `TRUSTED_IPS`/`TRUSTED_CIDRS`), the extractor returns it and ignores the headers, because the client connected directly. Otherwise each header is walked right-to-left; malformed entries are ignored, and the first entry that is not a trusted IP is returned. When every entry is trusted, the extractor returns the original requester, the first entry of the first header that has one. This matches Echo's `RealIP` semantics. Otherwise it returns `RemoteAddr`. The chain supports the special keywords `Direct` and `RemoteAddr` to force extraction from `http.Request.RemoteAddr`.

  `IP_EXTRACT_CONFIRMATION_HEADERS` adds an optional confirmation gate:
  - The value is a comma-separated list of header names.
  - A bare `HeaderName`, or `HeaderName=` with an empty value, requires presence.
  - `HeaderName=secret` requires a constant-time value match.
  - All entries must use the same mode; a mixed list is invalid.
  - The gate passes when at least one requirement passes.
  - A missing header, wrong value, mixed list, or invalid name/value returns `RemoteAddr`.
  - Whitespace around entries, names, and values is trimmed.
  - Confirmation header names and values are not logged.
  - Unset or empty preserves the current behavior.

  Example: a request arrives with `X-Forwarded-For: 127.0.0.1` and `RemoteAddr: 172.17.0.1:54321`. Both addresses are trusted, so the extractor returns `127.0.0.1`, not the proxy address.

  > **Security:** The headers are read only from trusted peers: loopback, private, link-local, or in `TRUSTED_IPS`/`TRUSTED_CIDRS`. A client on an untrusted network cannot choose the reported address. Right-to-left traversal skips values a client prepends when a trusted proxy appends the real client last. A proxy with a public address must be listed in `TRUSTED_CIDRS`, or the extractor reports the proxy address. Over-broad `TRUSTED_CIDRS` is the risk: the extractor would trust a range that includes untrusted clients. Behind an L4 proxy that cannot overwrite headers, set `IP_EXTRACT_CONFIRMATION_HEADERS` so requests that bypass the edge fail closed. Presence alone is forgeable; use `HeaderName=secret` when the origin is directly reachable.

  > **Proxy entries:** Do not add a private/internal proxy range to `TRUSTED_IPS` or `TRUSTED_CIDRS` unless you intend that proxy to be a trusted panic operator. The extractor already trusts loopback, private, and link-local peers, so the entry does not change header trust. It does make `PanicHandler` trust the proxy address: when the confirmation gate fails, `Execute` returns the proxy's `RemoteAddr`, and a listed proxy then unlocks full panic details. A public proxy must be listed for header extraction. If you also configure the confirmation gate, make sure it passes on that path, or a failed gate will return the proxy address and treat it as a trusted operator.

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

- **PaginationParser**: Parse pagination parameters like pageNumber, itemsPerPage, lastSeenId, sortBy, and sortDirection from HTTP requests. A zero itemsPerPage fails with `InvalidItemsPerPage`; a fractional pageNumber or itemsPerPage fails with `InvalidPageNumber` or `InvalidItemsPerPage`.

  ```go
  defaultPagination := tkDto.Pagination{PageNumber: 0, ItemsPerPage: 10}
  untrustedInput := map[string]any{"pageNumber": 1, "itemsPerPage": 20}
  parsedPagination, paginationParsingErr := PaginationParser(
    defaultPagination, untrustedInput,
  )
  ```

- **PaginationQueryParamsReader**: Read pagination query params and caller-listed optional params from an Echo context into the untrusted input map that `PaginationParser` consumes. It reads `pageNumber`, `itemsPerPage`, `sortBy`, `sortDirection`, and `lastSeenId`. A pagination value enters the map only when its tk conversion succeeds, so a malformed value stays out and `PaginationParser` returns the default. A zero `itemsPerPage` is dropped. Optional params enter the map when present and non-empty, with no conversion; a pagination key listed among them is ignored, so the converted value stands. The reader writes no defaults; the liaison owns them through `PaginationParser`.

  ```go
  optionalParamNames := []string{"nature", "name", "type", "status"}
  untrustedInput := tkPresentation.PaginationQueryParamsReader(
    echoContext, optionalParamNames,
  )
  ```

- **UiRequestInputReader**: Read DataTable/Carousel fragment query params into the liaison input map. It reads configured filters by `Key` and `QueryParamName`, with scalar, multi-select, and range kinds; range filters use `Min`/`Max` suffixes and return `{min, max}`. It composes `PaginationQueryParamsReader` for the unprefixed pagination keys, copies `BaseInput` shallowly, and injects `operatorAccountId`, `operatorSri`, and `operatorIpAddress` from the Echo context after reading query params. The fragment route identifies the component, so there is no component prefix. A missing or wrong-type operator account ID fails with `OperatorAccountIdNotFound`; query-supplied operator fields are overwritten. Pagination defaults stay with the liaison factory.

  ```go
  reader := tkPresentation.UiRequestInputReader{
    Filters: []tkPresentation.UiQueryParamFilter{
      {Key: "status"},
      {
        Key:            "tags",
        QueryParamName: "tag",
        Kind:           tkPresentation.UiQueryParamFilterKindMulti,
      },
      {
        Key:            "cpu",
        QueryParamName: "cores",
        Kind:           tkPresentation.UiQueryParamFilterKindRange,
      },
    },
    BaseInput: map[string]any{"withMetrics": true},
  }
  input, readingErr := reader.Read(echoContext)
  ```

  Map DataTable/Carousel kinds to the reader kinds: text and enum filters are scalar, multi-enum filters are multi, and number/date ranges are range. Add a scalar filter for the search placeholder when the table has a search box. The `QueryUrlTemplate` must map `{sortKey}` to `sortBy` for the pagination reader. Filters and pagination override `BaseInput`; operator fields always win.

- **ParseFieldMatchPatterns**: Parse a semicolon-separated list of `field:pattern` pairs into a map keyed by `FieldName`. Each entry is trimmed. An empty pattern after the colon defaults to `equals`. Field names keep their case. A missing colon fails with `InvalidFieldMatchPattern`; an invalid field name returns the `FieldName` constructor error (`FieldNameCannotBeEmpty`, `FieldNameTooBig`, or `InvalidFieldName`); an input that cannot convert to a string fails with `FieldMatchPatternsMustBeString`.

  ```go
  fieldMatchPatterns, parsingErr := ParseFieldMatchPatterns(
    "name:contains;status:equals",
  )
  ```

- **StringSliceVoParser**: Convert comma-separated, semicolon-separated, or array strings into value object slices.

  ```go
  rawInput := "note1,note2;note3"
  parsedNotes := StringSliceValueObjectParser(rawInput, tkValueObject.NewGenericNotes)
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

  err := errors.New("AccountNotFound")

  liaisonResponseNoMessage := NewLiaisonResponseNoMessage(
    LiaisonResponseStatusSuccess, err.Error(),
  )

  liaisonApiEmissionErr := LiaisonApiResponseEmitter(echoContext, liaisonResponse)

  LiaisonCliResponseRenderer(liaisonResponse)
  ```
