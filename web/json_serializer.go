package web

import (
	"bytes"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v5"
)

const maxPooledJSONBuf = 1 << 16 // 64 KiB

var jsonBufPool = sync.Pool{New: newJSONBuf}

func newJSONBuf() any { return new(bytes.Buffer) }

type JSONSerializer struct{}

func (s *JSONSerializer) Serialize(c *echo.Context, target any, indent string) error {
	enc := sonic.ConfigFastest.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(target)
}

func (s *JSONSerializer) Deserialize(c *echo.Context, target any) error {
	buf := jsonBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPooledJSONBuf {
			jsonBufPool.Put(buf)
		}
	}()

	if _, err := buf.ReadFrom(c.Request().Body); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	if err := sonic.Unmarshal(buf.Bytes(), target); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}
