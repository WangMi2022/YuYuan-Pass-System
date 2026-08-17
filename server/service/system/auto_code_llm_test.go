package system

import (
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common"
)

func TestBuildLLMAutoPathValidatesMode(t *testing.T) {
	original := global.GVA_CONFIG.AutoCode.AiPath
	t.Cleanup(func() {
		global.GVA_CONFIG.AutoCode.AiPath = original
	})
	global.GVA_CONFIG.AutoCode.AiPath = "https://ai.example.test/v1/{FUNC}"

	path, err := buildLLMAutoPath(common.JSONMap{"mode": "analysisChat"})
	if err != nil {
		t.Fatalf("build valid AI path: %v", err)
	}
	if path != "https://ai.example.test/v1/analysisChat" {
		t.Fatalf("path = %q, want allowlisted mode substitution", path)
	}

	for _, mode := range []string{"../admin", "/admin", "analysisChat?x=1", "analysisChat#fragment", ""} {
		_, err = buildLLMAutoPath(common.JSONMap{"mode": mode})
		if err == nil {
			t.Fatalf("mode %q should be rejected", mode)
		}
	}
}

func TestBuildLLMAutoPathRequiresConfiguration(t *testing.T) {
	original := global.GVA_CONFIG.AutoCode.AiPath
	t.Cleanup(func() {
		global.GVA_CONFIG.AutoCode.AiPath = original
	})
	global.GVA_CONFIG.AutoCode.AiPath = ""

	if _, err := buildLLMAutoPath(common.JSONMap{"mode": "ai"}); err == nil {
		t.Fatal("expected empty AI path to be rejected")
	}
}
