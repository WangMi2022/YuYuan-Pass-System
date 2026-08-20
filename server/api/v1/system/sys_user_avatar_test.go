package system

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	exampleModel "github.com/WangMi2022/mit-assets-admin/server/model/example"
	userModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestAttachUserAvatarPreviewDoesNotExposeManagedStorageOrigin(t *testing.T) {
	previousDB := global.GVA_DB
	previousConfig := global.GVA_CONFIG
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open avatar preview database: %v", err)
	}
	if err = database.AutoMigrate(&exampleModel.ExaFileUploadAndDownload{}); err != nil {
		t.Fatalf("migrate avatar media table: %v", err)
	}
	global.GVA_DB = database
	global.GVA_CONFIG.System.OssType = "minio"
	global.GVA_CONFIG.Minio.BucketUrl = "http://172.30.3.135:9000/gva-assets"
	global.GVA_CONFIG.JWT.SigningKey = "avatar-preview-test-key"
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		global.GVA_CONFIG = previousConfig
	})

	const key = "assets/2026-07-16/user-avatar.jpg"
	const rawURL = "http://172.30.3.135:9000/gva-assets/" + key
	if err = database.Create(&exampleModel.ExaFileUploadAndDownload{Name: "user-avatar.jpg", Key: key, Url: rawURL}).Error; err != nil {
		t.Fatalf("create avatar media record: %v", err)
	}

	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", "/user/getUserInfo", nil)
	user := userModel.SysUser{HeaderImg: rawURL}
	attachUserAvatarPreview(context, &user)

	if strings.Contains(user.HeaderImgPreviewURL, "172.30.3.135") || !strings.HasPrefix(user.HeaderImgPreviewURL, "/api/fileUploadAndDownload/preview?") {
		t.Fatalf("avatar response exposed storage URL: %q", user.HeaderImgPreviewURL)
	}
}
