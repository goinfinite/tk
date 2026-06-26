package tkPresentation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
)

const streamFloorBytes int64 = 5 * 1024 * 1024
const streamChunkSize int = 8192

func computeStreamCap(
	maxStreamSize int64,
	rng *rand.Rand,
) int64 {
	if maxStreamSize <= streamFloorBytes {
		return streamFloorBytes
	}
	rangeSize := maxStreamSize - streamFloorBytes
	return streamFloorBytes + rng.Int63n(rangeSize)
}

func (middleware *HoneypotMiddleware) serveStreamFallback(
	echoContext echo.Context,
) error {
	fallbackBody, _ := json.Marshal(map[string]string{
		"status":  "ok",
		"message": "Stream initialized",
	})
	echoContext.Response().Header().Set(
		"Content-Type", "application/json",
	)
	return echoContext.String(
		http.StatusOK, string(fallbackBody),
	)
}

func (middleware *HoneypotMiddleware) streamBandwidthExhaust(
	echoContext echo.Context,
) error {
	flusher, isFlusherAvailable :=
		echoContext.Response().Writer.(http.Flusher)
	if !isFlusherAvailable {
		return middleware.serveStreamFallback(echoContext)
	}
	responseWriter := echoContext.Response().Writer
	echoContext.Response().Header().Set(
		"Content-Type", "text/plain; charset=utf-8",
	)
	echoContext.Response().Header().Set(
		"Transfer-Encoding", "chunked",
	)
	responseWriter.WriteHeader(http.StatusOK)
	flusher.Flush()
	rng := rand.New(rand.NewSource(
		rand.Int63(),
	))
	streamCap := computeStreamCap(
		middleware.settings.MaxStreamSizeBytes.Int64(),
		rng,
	)
	totalWritten := int64(0)
	garbageLine := bandwidthGarbageLine()
	for totalWritten < streamCap {
		if echoContext.Request().Context().Err() != nil {
			return nil
		}
		chunk := buildBandwidthChunk(
			garbageLine, totalWritten,
		)
		written, writeErr := responseWriter.Write(chunk)
		if writeErr != nil {
			return nil
		}
		totalWritten += int64(written)
		flusher.Flush()
	}
	return nil
}

func bandwidthGarbageLine() string {
	return "DATA_BLOCK:" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" +
		"cccccccccccccccccccccccccccccccccccccccc" +
		"dddddddddddddddddddddddddddddddddddddddd" +
		"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" +
		"ffffffffffffffffffffffffffffffffffffffff" +
		"gggggggggggggggggggggggggggggggggggggggg\n"
}

func buildBandwidthChunk(
	garbageLine string,
	offset int64,
) []byte {
	header := fmt.Sprintf(
		"# chunk-offset: %d\n", offset,
	)
	chunk := make([]byte, 0, streamChunkSize)
	chunk = append(chunk, []byte(header)...)
	for len(chunk) < streamChunkSize {
		chunk = append(chunk, []byte(garbageLine)...)
	}
	return chunk[:streamChunkSize]
}

func (middleware *HoneypotMiddleware) streamAiTrap(
	echoContext echo.Context,
	interceptPath string,
) error {
	flusher, isFlusherAvailable :=
		echoContext.Response().Writer.(http.Flusher)
	if !isFlusherAvailable {
		return middleware.serveStreamFallback(echoContext)
	}
	responseWriter := echoContext.Response().Writer
	trapGenerator := honeypotAiTrapGenerator{}
	contentType := trapGenerator.contentType(interceptPath)
	echoContext.Response().Header().Set(
		"Content-Type", contentType,
	)
	echoContext.Response().Header().Set(
		"Transfer-Encoding", "chunked",
	)
	responseWriter.WriteHeader(http.StatusOK)
	flusher.Flush()
	rng := rand.New(rand.NewSource(
		rand.Int63(),
	))
	streamCap := computeStreamCap(
		middleware.settings.MaxStreamSizeBytes.Int64(),
		rng,
	)
	totalWritten := int64(0)
	chunkIndex := 0
	for totalWritten < streamCap {
		if echoContext.Request().Context().Err() != nil {
			return nil
		}
		chunk := trapGenerator.generateChunk(
			interceptPath, chunkIndex,
		)
		written, writeErr := responseWriter.Write(
			[]byte(chunk),
		)
		if writeErr != nil {
			return nil
		}
		totalWritten += int64(written)
		flusher.Flush()
		chunkIndex++
	}
	return nil
}
