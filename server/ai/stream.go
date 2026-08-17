package ai

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"hash"
	"io"
	"strings"
	"sync"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/goccy/go-json"
)

const maxStreamOutputBytes = 8 << 20

type streamAuditBody struct {
	io.ReadCloser
	state         *invocationState
	provider      string
	inputEstimate int64
	buffer        bytes.Buffer
	hasher        hash.Hash
	totalBytes    int64
	failed        error
	sawEOF        bool
	finished      bool
	mu            sync.Mutex
}

func (b *streamAuditBody) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	remaining := int64(maxStreamOutputBytes) - b.totalBytes
	if remaining == 0 {
		var probe [1]byte
		n, err := b.ReadCloser.Read(probe[:])
		if n > 0 {
			b.failLimit()
			return 0, b.failed
		}
		if err == io.EOF {
			b.sawEOF = true
			b.complete()
		} else if err != nil {
			b.failed = &Error{Type: ErrorTypeProvider, Message: "读取模型流式响应失败", Cause: err}
		}
		return 0, err
	}
	readSize := len(data)
	if int64(readSize) > remaining+1 {
		readSize = int(remaining + 1)
	}
	n, err := b.ReadCloser.Read(data[:readSize])
	accepted := n
	if int64(accepted) > remaining {
		accepted = int(remaining)
	}
	if n > 0 {
		b.totalBytes += int64(accepted)
		_, _ = b.hasher.Write(data[:accepted])
		_, _ = b.buffer.Write(data[:accepted])
	}
	if accepted < n {
		b.failLimit()
		return accepted, b.failed
	}
	if err != nil && err != io.EOF {
		b.failed = &Error{Type: ErrorTypeProvider, Message: "读取模型流式响应失败", Cause: err}
	}
	if err == io.EOF {
		b.sawEOF = true
		b.complete()
	}
	return n, err
}

func (b *streamAuditBody) Close() error {
	err := b.ReadCloser.Close()
	if err != nil && b.failed == nil {
		b.failed = &Error{Type: ErrorTypeProvider, Message: "关闭模型流式响应失败", Cause: err}
	} else if !b.sawEOF && b.failed == nil {
		b.failed = &Error{Type: ErrorTypeProvider, Message: "模型流式响应未完整读取"}
	}
	b.complete()
	return err
}

func (b *streamAuditBody) failLimit() {
	b.failed = &Error{Type: ErrorTypeSchema, Message: "模型流式输出超过 8MB 限制"}
	_ = b.ReadCloser.Close()
	b.complete()
}

func (b *streamAuditBody) complete() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.finished {
		return
	}
	b.finished = true
	if b.failed != nil {
		b.state.fail(b.failed)
		b.state.finish()
		return
	}
	inputTokens, outputTokens := parseStreamUsage(b.provider, b.buffer.Bytes())
	if inputTokens == 0 {
		inputTokens = b.inputEstimate
	}
	if outputTokens == 0 {
		outputTokens = estimateTokens(b.totalBytes)
	}
	b.state.success(inputTokens, outputTokens, hex.EncodeToString(b.hasher.Sum(nil)))
	b.state.finish()
}

func parseStreamUsage(providerName string, body []byte) (int64, int64) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), maxStreamOutputBytes)
	var inputTokens, outputTokens int64
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		if providerName == config.AIProviderAnthropic {
			var event struct {
				Usage struct {
					InputTokens  int64 `json:"input_tokens"`
					OutputTokens int64 `json:"output_tokens"`
				} `json:"usage"`
				Message struct {
					Usage struct {
						InputTokens  int64 `json:"input_tokens"`
						OutputTokens int64 `json:"output_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if json.Unmarshal([]byte(payload), &event) == nil {
				inputTokens = maxInt64(inputTokens, maxInt64(event.Usage.InputTokens, event.Message.Usage.InputTokens))
				outputTokens = maxInt64(outputTokens, maxInt64(event.Usage.OutputTokens, event.Message.Usage.OutputTokens))
			}
			continue
		}
		var event struct {
			Usage struct {
				PromptTokens     int64 `json:"prompt_tokens"`
				CompletionTokens int64 `json:"completion_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal([]byte(payload), &event) == nil {
			inputTokens = maxInt64(inputTokens, event.Usage.PromptTokens)
			outputTokens = maxInt64(outputTokens, event.Usage.CompletionTokens)
		}
	}
	return inputTokens, outputTokens
}

func estimateTokens(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	return (bytes + 3) / 4
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
