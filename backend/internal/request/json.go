package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	if len(bytes.TrimSpace(body)) == 0 || bytes.TrimSpace(body)[0] != '{' {
		response.Fail(c, 400, "invalid_json", "JSON object required")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		var large *http.MaxBytesError
		if errors.As(err, &large) {
			response.Fail(c, 413, "body_too_large", "request body exceeds 64 KiB")
		} else {
			response.Fail(c, 400, "invalid_json", "valid JSON object required")
		}
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		response.Fail(c, 400, "invalid_json", "exactly one JSON object required")
		return false
	}
	if err := binding.Validator.ValidateStruct(out); err != nil {
		response.Fail(c, 422, "validation_failed", "fields are missing or invalid")
		return false
	}
	return true
}
