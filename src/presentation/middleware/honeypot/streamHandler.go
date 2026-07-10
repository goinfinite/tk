package tkPresentationMiddlewareHoneypot

import (
	"crypto/rand"
	"math/big"
	"net/http"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

const (
	streamMinSizeBytes uint64 = 5 * 1024 * 1024
	streamMaxSizeBytes uint64 = 20 * 1024 * 1024
)

type StreamHandler struct {
	maxStreamSize tkValueObject.HoneypotMaxStreamSize
}

func NewStreamHandler(
	maxStreamSize tkValueObject.HoneypotMaxStreamSize,
) *StreamHandler {
	return &StreamHandler{maxStreamSize: maxStreamSize}
}

func (handler *StreamHandler) ServeBandwidthExhaust(
	echoContext echo.Context,
) error {
	streamSize := handler.resolveStreamSize()
	if streamSize == 0 {
		return echoContext.NoContent(http.StatusOK)
	}

	echoContext.Response().Header().Set(
		echo.HeaderContentType, "application/octet-stream",
	)
	echoContext.Response().WriteHeader(http.StatusOK)

	return handler.writeRandomBytes(echoContext, streamSize)
}

func (handler *StreamHandler) ServeAiTrap(
	echoContext echo.Context,
	generator *AiTrapGenerator,
) error {
	maxSize := handler.maxStreamSize.Uint64()
	if maxSize == 0 {
		return echoContext.NoContent(http.StatusOK)
	}

	echoContext.Response().Header().Set(
		echo.HeaderContentType, "text/plain; charset=utf-8",
	)
	echoContext.Response().WriteHeader(http.StatusOK)

	generatedText := generator.Generate(int(maxSize))
	_, writeErr := echoContext.Response().Write([]byte(generatedText))
	return writeErr
}

func (handler *StreamHandler) resolveStreamSize() uint64 {
	maxAllowed := handler.maxStreamSize.Uint64()
	if maxAllowed == 0 {
		return 0
	}

	rangeSize := streamMaxSizeBytes - streamMinSizeBytes + 1
	randomOffset, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
	if err != nil {
		return streamMinSizeBytes
	}

	randomSize := streamMinSizeBytes + uint64(randomOffset.Int64())
	if randomSize > maxAllowed {
		return maxAllowed
	}

	return randomSize
}

func (handler *StreamHandler) writeRandomBytes(
	echoContext echo.Context,
	totalBytes uint64,
) error {
	chunkSize := 32 * 1024
	chunk := make([]byte, chunkSize)

	var written uint64
	for written < totalBytes {
		currentChunkSize := uint64(chunkSize)
		remaining := totalBytes - written
		if remaining < currentChunkSize {
			currentChunkSize = remaining
		}

		_, readErr := rand.Read(chunk[:currentChunkSize])
		if readErr != nil {
			return readErr
		}

		_, writeErr := echoContext.Response().Write(chunk[:currentChunkSize])
		if writeErr != nil {
			return writeErr
		}

		written += currentChunkSize
	}

	return nil
}
