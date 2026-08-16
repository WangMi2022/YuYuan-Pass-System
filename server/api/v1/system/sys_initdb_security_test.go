package system

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
)

func TestAllowDatabaseInitialization(t *testing.T) {
	previousDB := global.GVA_DB
	previousTokenConsumed := installTokenConsumed
	global.GVA_DB = nil
	installTokenConsumed = false
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		installTokenConsumed = previousTokenConsumed
	})
	t.Setenv("GVA_INSTALL_TOKEN", "install-token-for-test")

	request := func(remoteAddr string) *gin.Context {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		httpRequest := httptest.NewRequest(http.MethodPost, "/init/initdb", nil)
		httpRequest.RemoteAddr = remoteAddr
		httpRequest.Header.Set("X-Install-Token", "install-token-for-test")
		context.Request = httpRequest
		return context
	}

	t.Setenv("GVA_INSTALL_MODE", "false")
	if allowDatabaseInitialization(request("127.0.0.1:50000")) {
		t.Fatal("initialization must require explicit install mode")
	}

	t.Setenv("GVA_INSTALL_MODE", "true")
	if allowDatabaseInitialization(request("10.0.0.10:50000")) {
		t.Fatal("initialization must reject non-loopback clients")
	}
	if !allowDatabaseInitialization(request("127.0.0.1:50000")) {
		t.Fatal("initialization must accept a valid local installation token")
	}

	installTokenConsumed = true
	if allowDatabaseInitialization(request("127.0.0.1:50000")) {
		t.Fatal("initialization must reject a consumed installation token")
	}
}
