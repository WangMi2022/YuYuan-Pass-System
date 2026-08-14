package ai

import "testing"

func TestOpenAIChatURL(t *testing.T) {
	cases := map[string]string{
		"https://example.com":                     "https://example.com/v1/chat/completions",
		"https://example.com/v1":                  "https://example.com/v1/chat/completions",
		"https://example.com/v1/chat/completions": "https://example.com/v1/chat/completions",
	}
	for input, expected := range cases {
		if actual := openAIChatURL(input); actual != expected {
			t.Fatalf("openAIChatURL(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestOpenAIContent(t *testing.T) {
	content, err := openAIContent([]byte(`"ready"`))
	if err != nil || content != "ready" {
		t.Fatalf("content = %q, err = %v", content, err)
	}
}
