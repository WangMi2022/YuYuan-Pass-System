package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
)

func TestMultimodalRecognizerRetriesOversizedOpenAIImage(t *testing.T) {
	requestBodies := make([][]byte, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		requestBodies = append(requestBodies, body)
		if len(requestBodies) == 1 {
			response.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		mimeType, imageData := openAIImageFromRequest(t, body)
		if mimeType != "image/jpeg" {
			t.Fatalf("retry image MIME = %q, want image/jpeg", mimeType)
		}
		if _, _, err = image.DecodeConfig(bytes.NewReader(imageData)); err != nil {
			t.Fatalf("retry image cannot be decoded: %v", err)
		}
		writeOpenAIInvoiceSuccess(response, "413-RETRIED")
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, APIKey: "test-key", Model: "vision-model",
		Protocol: config.MultimodalProtocolOpenAICompatible,
		Timeout:  time.Second, AllowPrivateEndpoints: true,
	}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "large-invoice.jpg", ContentType: "image/jpeg", Data: noisyJPEG(t, 1800, 1300),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "413-RETRIED" {
		t.Fatalf("invoice number = %q", result.InvoiceNumber)
	}
	if len(requestBodies) != 2 {
		t.Fatalf("requests = %d, want 2", len(requestBodies))
	}
	if len(requestBodies[0]) > 900<<10 {
		t.Fatalf("first body = %d bytes, want at most 900 KiB", len(requestBodies[0]))
	}
	if len(requestBodies[1]) >= len(requestBodies[0]) {
		t.Fatalf("retry body = %d bytes, first body = %d bytes", len(requestBodies[1]), len(requestBodies[0]))
	}
	if len(requestBodies[1]) > 576<<10 {
		t.Fatalf("retry body = %d bytes, want at most 576 KiB", len(requestBodies[1]))
	}
}

func TestMultimodalRecognizerReencodesSmallImageAfterOpenAI413(t *testing.T) {
	requestBodies := make([][]byte, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		requestBodies = append(requestBodies, body)
		if len(requestBodies) == 1 {
			response.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		mimeType, _ := openAIImageFromRequest(t, body)
		if mimeType != "image/jpeg" {
			t.Fatalf("retry image MIME = %q, want image/jpeg", mimeType)
		}
		writeOpenAIInvoiceSuccess(response, "SMALL-413-RETRIED")
	}))
	defer server.Close()

	imageData := noisyPNG(t, 360, 280)
	if len(imageData) >= maxMultimodalRetryImageSize {
		t.Fatalf("test image = %d bytes, want less than retry threshold", len(imageData))
	}
	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, Model: "vision-model",
		Protocol: config.MultimodalProtocolOpenAICompatible,
		Timeout:  time.Second, AllowPrivateEndpoints: true,
	}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "small-invoice.png", ContentType: "image/png", Data: imageData,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "SMALL-413-RETRIED" || len(requestBodies) != 2 {
		t.Fatalf("result = %#v, requests = %d", result, len(requestBodies))
	}
	if len(requestBodies[1]) >= len(requestBodies[0]) {
		t.Fatalf("retry body = %d bytes, first body = %d bytes", len(requestBodies[1]), len(requestBodies[0]))
	}
}

func TestMultimodalRecognizerAppliesJPEGOrientationBeforeCompression(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		_, imageData := openAIImageFromRequest(t, body)
		prepared, _, err := image.Decode(bytes.NewReader(imageData))
		if err != nil {
			t.Fatal(err)
		}
		if prepared.Bounds().Dx() >= prepared.Bounds().Dy() {
			t.Fatalf("prepared dimensions = %dx%d, want portrait after Orientation 6", prepared.Bounds().Dx(), prepared.Bounds().Dy())
		}
		assertCornerColor(t, prepared, "top-left", prepared.Bounds().Dx()/12, prepared.Bounds().Dy()/12, color.RGBA{B: 255, A: 255})
		assertCornerColor(t, prepared, "top-right", prepared.Bounds().Dx()*11/12, prepared.Bounds().Dy()/12, color.RGBA{R: 255, A: 255})
		writeOpenAIInvoiceSuccess(response, "ORIENTED")
	}))
	defer server.Close()

	imageData := orientedNoisyJPEG(t, 1600, 900, 6)
	if len(imageData) <= maxMultimodalInitialImageSize {
		t.Fatalf("test image = %d bytes, want compression path", len(imageData))
	}
	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, Model: "vision-model",
		Protocol: config.MultimodalProtocolOpenAICompatible,
		Timeout:  time.Second, AllowPrivateEndpoints: true,
	}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "phone-invoice.jpg", ContentType: "image/jpeg", Data: imageData,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "ORIENTED" {
		t.Fatalf("invoice number = %q", result.InvoiceNumber)
	}
}

func TestMultimodalRecognizerRetriesOpenAI413OnlyOnce(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		response.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, Model: "vision-model",
		Protocol: config.MultimodalProtocolOpenAICompatible,
		Timeout:  time.Second, AllowPrivateEndpoints: true,
	}
	_, err := recognizer.Recognize(context.Background(), Input{
		FileName: "large-invoice.jpg", ContentType: "image/jpeg", Data: noisyJPEG(t, 1300, 900),
	})
	if err == nil || !strings.Contains(err.Error(), "系统压缩后仍被拒绝") {
		t.Fatalf("error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requests = %d, want 2", requestCount)
	}
}

func TestMultimodalRecognizerDoesNotRetryNon413Response(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		response.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, Model: "vision-model",
		Protocol: config.MultimodalProtocolOpenAICompatible,
		Timeout:  time.Second, AllowPrivateEndpoints: true,
	}
	_, err := recognizer.Probe(context.Background())
	if err == nil || !strings.Contains(err.Error(), "状态 502") {
		t.Fatalf("error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("requests = %d, want 1", requestCount)
	}
}

func TestMultimodalRecognizerStopsProtocolDetectionAfter413Retry(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		response.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, Model: "vision-model",
		Timeout: time.Second, AllowPrivateEndpoints: true,
	}
	_, err := recognizer.Recognize(context.Background(), Input{
		FileName: "large-invoice.jpg", ContentType: "image/jpeg", Data: noisyJPEG(t, 1300, 900),
	})
	if err == nil || !strings.Contains(err.Error(), "系统压缩后仍被拒绝") {
		t.Fatalf("error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requests = %d, want 2", requestCount)
	}
}

func TestMultimodalRecognizerStopsProtocolDetectionAfter502(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount++
		response.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, Model: "vision-model",
		Timeout: time.Second, AllowPrivateEndpoints: true,
	}
	_, err := recognizer.Probe(context.Background())
	if err == nil || !strings.Contains(err.Error(), "状态 502") {
		t.Fatalf("error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("requests = %d, want 1", requestCount)
	}
}

func TestMultimodalRecognizerUsesImageInputAndParsesResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		serialized, _ := json.Marshal(body)
		if !strings.Contains(string(serialized), "data:image/png;base64,") {
			t.Fatalf("request does not contain an image data URL: %s", serialized)
		}
		content := `{"invoiceType":"电子发票","invoiceNumber":"12345678","issueDate":"2026-07-29","sellerName":"示例公司","amountCents":1000,"taxCents":60,"totalCents":1060,"rawText":"发票原文","confidence":0.91,"fieldConfidences":{"invoiceNumber":0.96},"items":[]}`
		_ = json.NewEncoder(response).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "```json\n" + content + "\n```"}}},
		})
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{BaseURL: server.URL, APIKey: "test-key", Model: "vision-model", Timeout: time.Second, AllowPrivateEndpoints: true}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "invoice.png", ContentType: "image/png", Data: probePNG,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "12345678" || result.TotalCents != 1060 || result.IssueDate == nil {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestChatCompletionsURLAcceptsBaseOrFullEndpoint(t *testing.T) {
	tests := map[string]string{
		"https://api.example.com":                     "https://api.example.com/v1/chat/completions",
		"https://api.example.com/v1/":                 "https://api.example.com/v1/chat/completions",
		"https://api.example.com/v1/chat/completions": "https://api.example.com/v1/chat/completions",
		"https://api.example.com/v1?tenant=asset":     "https://api.example.com/v1/chat/completions?tenant=asset",
	}
	for input, expected := range tests {
		if actual := chatCompletionsURL(input); actual != expected {
			t.Fatalf("%s: got %s want %s", input, actual, expected)
		}
	}
}

func TestMultimodalRecognizerDetectsAnthropicProtocol(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		switch request.URL.Path {
		case "/v1/chat/completions":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]any{
					"content": "", "reasoning_content": "thinking",
				}}},
			})
		case "/v1/messages":
			if request.Header.Get("x-api-key") != "test-key" {
				t.Fatal("missing Anthropic API key header")
			}
			if request.Header.Get("anthropic-version") != "2023-06-01" {
				t.Fatal("missing Anthropic version header")
			}
			if request.Header.Get("Authorization") != "" {
				t.Fatal("Anthropic request should not use the Bearer header")
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			serialized, _ := json.Marshal(body)
			if !strings.Contains(string(serialized), `"type":"base64"`) ||
				!strings.Contains(string(serialized), `"media_type":"image/png"`) {
				t.Fatalf("request does not contain an Anthropic image source: %s", serialized)
			}
			if body["max_tokens"] != float64(probeMaxTokens) {
				t.Fatalf("probe max_tokens = %#v, want %d", body["max_tokens"], probeMaxTokens)
			}
			_ = json.NewEncoder(response).Encode(map[string]any{
				"content": []map[string]any{{"type": "thinking", "thinking": "done"}, {"type": "text", "text": "OK"}},
			})
		default:
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{BaseURL: server.URL, APIKey: "test-key", Model: "vision-model", Timeout: time.Second, AllowPrivateEndpoints: true}
	protocol, err := recognizer.Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if protocol != config.MultimodalProtocolAnthropic || requestCount != 2 {
		t.Fatalf("detected protocol = %q, requests = %d", protocol, requestCount)
	}
}

func TestMultimodalRecognizerUsesStoredAnthropicProtocol(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		content := `{"invoiceType":"电子发票","invoiceNumber":"87654321","issueDate":"2026-07-29","sellerName":"示例公司","totalCents":1060,"rawText":"发票原文","confidence":0.93,"fieldConfidences":{},"items":[]}`
		_ = json.NewEncoder(response).Encode(map[string]any{
			"content": []map[string]any{{"type": "text", "text": content}},
		})
	}))
	defer server.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: server.URL, APIKey: "test-key", Model: "vision-model",
		Protocol: config.MultimodalProtocolAnthropic, Timeout: time.Second, AllowPrivateEndpoints: true,
	}
	result, err := recognizer.Recognize(context.Background(), Input{
		FileName: "invoice.png", ContentType: "image/png", Data: probePNG,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InvoiceNumber != "87654321" || requestCount != 1 {
		t.Fatalf("unexpected result or request count: %#v, %d", result, requestCount)
	}
}

func TestAnthropicMessagesURLAcceptsBaseOrFullEndpoint(t *testing.T) {
	tests := map[string]string{
		"https://api.example.com":                       "https://api.example.com/v1/messages",
		"https://api.example.com/v1/":                   "https://api.example.com/v1/messages",
		"https://api.example.com/anthropic":             "https://api.example.com/anthropic/v1/messages",
		"https://api.example.com/anthropic/v1/messages": "https://api.example.com/anthropic/v1/messages",
		"https://api.example.com/v1?tenant=asset":       "https://api.example.com/v1/messages?tenant=asset",
	}
	for input, expected := range tests {
		if actual := anthropicMessagesURL(input); actual != expected {
			t.Fatalf("%s: got %s want %s", input, actual, expected)
		}
	}
}

func TestMultimodalRecognizerDoesNotForwardAPIKeyThroughRedirect(t *testing.T) {
	destinationRequests := 0
	destination := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		destinationRequests++
		if request.Header.Get("x-api-key") != "" || request.Header.Get("Authorization") != "" {
			t.Errorf("provider credential was forwarded through redirect")
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer destination.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", request.URL.Path)
		}
		http.Redirect(response, request, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()

	recognizer := MultimodalRecognizer{
		BaseURL: redirect.URL, APIKey: "test-key", Model: "vision-model",
		Protocol: config.MultimodalProtocolAnthropic, Timeout: time.Second, AllowPrivateEndpoints: true,
	}
	if _, err := recognizer.Probe(context.Background()); err == nil {
		t.Fatal("redirect response was accepted")
	}
	if destinationRequests != 0 {
		t.Fatalf("redirect destination received %d requests", destinationRequests)
	}
}

func noisyJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := noisyImage(width, height)
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 96}); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func noisyPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, noisyImage(width, height)); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func orientedNoisyJPEG(t *testing.T, width, height, orientation int) []byte {
	t.Helper()
	img := noisyImage(width, height)
	paintCorner(img, image.Rect(0, 0, width/5, height/5), color.RGBA{R: 255, A: 255})
	paintCorner(img, image.Rect(width-width/5, 0, width, height/5), color.RGBA{G: 255, A: 255})
	paintCorner(img, image.Rect(0, height-height/5, width/5, height), color.RGBA{B: 255, A: 255})
	paintCorner(img, image.Rect(width-width/5, height-height/5, width, height), color.RGBA{R: 255, G: 255, A: 255})
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 96}); err != nil {
		t.Fatal(err)
	}
	data := encoded.Bytes()
	exif := []byte{
		0xff, 0xe1, 0x00, 0x22,
		'E', 'x', 'i', 'f', 0x00, 0x00,
		'I', 'I', 0x2a, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x01, 0x00,
		0x12, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	binary.LittleEndian.PutUint16(exif[28:30], uint16(orientation))
	withEXIF := make([]byte, 0, len(data)+len(exif))
	withEXIF = append(withEXIF, data[:2]...)
	withEXIF = append(withEXIF, exif...)
	withEXIF = append(withEXIF, data[2:]...)
	return withEXIF
}

func paintCorner(img *image.RGBA, bounds image.Rectangle, value color.RGBA) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.SetRGBA(x, y, value)
		}
	}
}

func assertCornerColor(t *testing.T, img image.Image, name string, x, y int, expected color.RGBA) {
	t.Helper()
	actual := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	const dominance = 80
	if expected.R > 0 && (int(actual.R) < int(actual.G)+dominance || int(actual.R) < int(actual.B)+dominance) {
		t.Fatalf("%s color = %#v, want red-dominant", name, actual)
	}
	if expected.B > 0 && (int(actual.B) < int(actual.R)+dominance || int(actual.B) < int(actual.G)+dominance) {
		t.Fatalf("%s color = %#v, want blue-dominant", name, actual)
	}
}

func noisyImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := uint8((x*31 + y*17 + x*y*13) % 256)
			img.SetRGBA(x, y, color.RGBA{
				R: value,
				G: uint8((int(value)*53 + x*7) % 256),
				B: uint8((int(value)*97 + y*11) % 256),
				A: 255,
			})
		}
	}
	return img
}

func openAIImageFromRequest(t *testing.T, body []byte) (string, []byte) {
	t.Helper()
	var request struct {
		Messages []struct {
			Content []struct {
				Type     string `json:"type"`
				ImageURL struct {
					URL string `json:"url"`
				} `json:"image_url"`
			} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	for _, message := range request.Messages {
		for _, content := range message.Content {
			if content.Type != "image_url" {
				continue
			}
			prefix, encoded, found := strings.Cut(content.ImageURL.URL, ";base64,")
			if !found || !strings.HasPrefix(prefix, "data:") {
				t.Fatalf("invalid image data URL")
			}
			data, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			return strings.TrimPrefix(prefix, "data:"), data
		}
	}
	t.Fatal("OpenAI request does not contain image data")
	return "", nil
}

func writeOpenAIInvoiceSuccess(response http.ResponseWriter, invoiceNumber string) {
	content := fmt.Sprintf(`{"invoiceType":"电子发票","invoiceNumber":%q,"issueDate":"2026-07-29","sellerName":"示例公司","totalCents":1060,"confidence":0.9,"fieldConfidences":{},"items":[]}`, invoiceNumber)
	_ = json.NewEncoder(response).Encode(map[string]any{
		"choices": []map[string]any{{"message": map[string]any{"content": content}}},
	})
}
