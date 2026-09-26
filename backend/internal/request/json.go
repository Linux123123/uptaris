package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func JSON(c *gin.Context, out any) bool {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			response.Fail(c, 413, "body_too_large", "request body exceeds 64 KiB")
		} else {
			response.Fail(c, 400, "invalid_json", "could not read request body")
		}
		return false
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		response.Fail(c, 400, "invalid_json", "JSON object required")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		response.Fail(c, 400, "invalid_json", decodeMessage(err))
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		response.Fail(c, 400, "invalid_json", "exactly one JSON object required")
		return false
	}
	return true
}

func decodeMessage(err error) string {
	var fieldType *json.UnmarshalTypeError
	if errors.As(err, &fieldType) && fieldType.Field != "" {
		return fmt.Sprintf("%s must be %s", fieldType.Field, fieldType.Type)
	}
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		return fmt.Sprintf("invalid JSON at byte %d", syntax.Offset)
	}
	if strings.HasPrefix(err.Error(), "json: unknown field ") {
		return err.Error()
	}
	return "valid JSON object required"
}
