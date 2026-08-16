package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/utils"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const bufferSize = 1024

const (
	fileBodySummary      = "[文件]"
	truncatedBodySummary = "[超出记录长度]"
	binaryBodySummary    = "[二进制响应]"
	redactedValue        = "[已脱敏]"
)

var sensitiveJSONValuePattern = regexp.MustCompile(`(?i)("(?:api[-_]?key|password|secret|token|authorization|auth[-_]?header|signing[-_]?key|access[-_]?key(?:[-_](?:id|secret))?|secret[-_]?(?:id|key))"\s*:\s*)"(?:\\.|[^"\\])*(?:"|$)`)

func OperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userId int
		requestBody := newCappedBuffer(bufferSize)
		isMultipart := strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "multipart/form-data")
		if c.Request.Method != http.MethodGet && !isMultipart && c.Request.Body != nil {
			originalBody := c.Request.Body
			c.Request.Body = &teeReadCloser{
				Reader: io.TeeReader(originalBody, requestBody),
				Closer: originalBody,
			}
		}
		claims, _ := utils.GetClaims(c)
		if claims != nil && claims.BaseClaims.ID != 0 {
			userId = int(claims.BaseClaims.ID)
		} else {
			id, err := strconv.Atoi(c.Request.Header.Get("x-user-id"))
			if err != nil {
				userId = 0
			}
			userId = id
		}
		record := system.SysOperationRecord{
			Ip:     c.ClientIP(),
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Agent:  c.Request.UserAgent(),
			UserID: userId,
		}

		if c.Request.Method == http.MethodGet {
			query, _ := json.Marshal(c.Request.URL.Query())
			_, _ = requestBody.Write(query)
		}

		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           newCappedBuffer(bufferSize),
		}
		c.Writer = writer
		now := time.Now()

		c.Next()

		latency := time.Since(now)
		record.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		record.Status = c.Writer.Status()
		record.Latency = latency
		if isMultipart {
			record.Body = fileBodySummary
		} else {
			record.Body = sanitizeOperationBody(requestBody.Summary())
		}
		if isBinaryResponse(c.Writer.Header()) {
			record.Resp = binaryBodySummary
		} else {
			record.Resp = sanitizeOperationBody(writer.body.Summary())
		}
		if err := global.GVA_DB.Create(&record).Error; err != nil {
			global.GVA_LOG.Error("create operation record error:", zap.Error(err))
		}
	}
}

func sanitizeOperationBody(body string) string {
	if body == "" {
		return body
	}
	truncated := strings.HasSuffix(body, truncatedBodySummary)
	payload := strings.TrimSuffix(body, truncatedBodySummary)
	var decoded any
	if !truncated && json.Unmarshal([]byte(payload), &decoded) == nil {
		redactSensitiveJSON(decoded)
		if sanitized, err := json.Marshal(decoded); err == nil {
			return string(sanitized)
		}
	}
	sanitized := sensitiveJSONValuePattern.ReplaceAllString(payload, `${1}"`+redactedValue+`"`)
	if truncated {
		sanitized += truncatedBodySummary
	}
	return sanitized
}

func redactSensitiveJSON(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if isSensitiveOperationField(key) {
				typed[key] = redactedValue
				continue
			}
			redactSensitiveJSON(nested)
		}
	case []any:
		for _, nested := range typed {
			redactSensitiveJSON(nested)
		}
	}
}

func isSensitiveOperationField(key string) bool {
	normalized := strings.NewReplacer("-", "", "_", "", " ", "").Replace(strings.ToLower(key))
	if strings.HasSuffix(normalized, "configured") {
		return false
	}
	return strings.Contains(normalized, "password") || strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") || strings.Contains(normalized, "apikey") ||
		normalized == "authorization" || normalized == "authheader" || normalized == "signingkey" ||
		normalized == "accesskey" || normalized == "accesskeyid" || normalized == "question" ||
		normalized == "prompt" || normalized == "answer" || normalized == "response" ||
		normalized == "content" || normalized == "rawtext"
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *cappedBuffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	_, _ = r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseBodyWriter) WriteString(s string) (int, error) {
	_, _ = r.body.Write([]byte(s))
	return r.ResponseWriter.WriteString(s)
}

type cappedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{limit: limit}
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	written := len(p)
	remaining := b.limit - b.Len()
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		_, _ = b.Buffer.Write(p[:remaining])
	}
	if written > remaining {
		b.truncated = true
	}
	return written, nil
}

func (b *cappedBuffer) Summary() string {
	prefix := strings.ToValidUTF8(b.String(), "�")
	if b.truncated {
		return prefix + truncatedBodySummary
	}
	return prefix
}

type teeReadCloser struct {
	io.Reader
	io.Closer
}

func isBinaryResponse(header http.Header) bool {
	contentType := strings.ToLower(header.Get("Content-Type"))
	contentDisposition := strings.ToLower(header.Get("Content-Disposition"))
	contentTransferEncoding := strings.ToLower(header.Get("Content-Transfer-Encoding"))
	return strings.Contains(contentDisposition, "attachment") ||
		strings.Contains(contentTransferEncoding, "binary") ||
		strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "video/") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentType, "application/pdf") ||
		strings.Contains(contentType, "application/zip") ||
		strings.Contains(contentType, "application/force-download") ||
		strings.Contains(contentType, "application/download") ||
		strings.Contains(contentType, "application/vnd.ms-excel") ||
		strings.Contains(contentType, "application/vnd.openxmlformats-officedocument")
}
