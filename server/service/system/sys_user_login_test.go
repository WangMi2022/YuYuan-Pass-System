package system

import (
	"fmt"
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	systemModel "github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserLoginTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB := global.GVA_DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open login test database: %v", err)
	}
	if err = db.AutoMigrate(
		&systemModel.SysAuthority{},
		&systemModel.SysUser{},
		&systemModel.SysAuthorityMenu{},
		&systemModel.SysBaseMenu{},
	); err != nil {
		t.Fatalf("migrate login tables: %v", err)
	}
	if err = db.Create(&systemModel.SysAuthority{
		AuthorityId:   888,
		AuthorityName: "管理员",
		DefaultRouter: "dashboard",
	}).Error; err != nil {
		t.Fatalf("seed login authority: %v", err)
	}
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func seedLoginUser(t *testing.T, db *gorm.DB, username, email, password string) systemModel.SysUser {
	t.Helper()
	user := systemModel.SysUser{
		UUID:        uuid.New(),
		Username:    username,
		Email:       email,
		Password:    utils.BcryptHash(password),
		NickName:    username,
		AuthorityId: 888,
		Enable:      1,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed login user %q: %v", username, err)
	}
	return user
}

func TestUserServiceLoginAcceptsUsernameAndUniqueEmail(t *testing.T) {
	db := setupUserLoginTestDB(t)
	want := seedLoginUser(t, db, "asset.admin", "Leader@Example.com", "correct-pass")
	service := new(UserService)

	byUsername, err := service.Login(&systemModel.SysUser{Username: "asset.admin", Password: "correct-pass"})
	if err != nil {
		t.Fatalf("login by username: %v", err)
	}
	if byUsername.ID != want.ID || byUsername.Username != want.Username {
		t.Fatalf("username login returned user ID=%d username=%q", byUsername.ID, byUsername.Username)
	}

	byEmail, err := service.Login(&systemModel.SysUser{Username: "  leader@EXAMPLE.com  ", Password: "correct-pass"})
	if err != nil {
		t.Fatalf("login by email: %v", err)
	}
	if byEmail.ID != want.ID || byEmail.Username != want.Username {
		t.Fatalf("email login returned user ID=%d username=%q", byEmail.ID, byEmail.Username)
	}
}

func TestUserServiceLoginRejectsAmbiguousEmail(t *testing.T) {
	db := setupUserLoginTestDB(t)
	seedLoginUser(t, db, "first.user", "shared@example.com", "first-pass")
	seedLoginUser(t, db, "second.user", "SHARED@example.com", "second-pass")

	_, err := new(UserService).Login(&systemModel.SysUser{
		Username: "shared@example.com",
		Password: "first-pass",
	})
	if err == nil || err.Error() != "邮箱对应多个账号，请使用用户名登录" {
		t.Fatalf("ambiguous email error = %v", err)
	}
}

func TestUserServiceLoginPrefersExactUsernameOverEmail(t *testing.T) {
	db := setupUserLoginTestDB(t)
	want := seedLoginUser(t, db, "owner@example.com", "owner-contact@example.com", "username-pass")
	seedLoginUser(t, db, "other.user", "owner@example.com", "email-pass")

	got, err := new(UserService).Login(&systemModel.SysUser{
		Username: "owner@example.com",
		Password: "username-pass",
	})
	if err != nil {
		t.Fatalf("login by exact username: %v", err)
	}
	if got.ID != want.ID {
		t.Fatalf("exact username login returned ID=%d, want %d", got.ID, want.ID)
	}
}

func TestUserServiceLoginPreservesExactWhitespaceUsername(t *testing.T) {
	db := setupUserLoginTestDB(t)
	want := seedLoginUser(t, db, " spaced.user ", "spaced@example.com", "spaced-pass")
	service := new(UserService)

	got, err := service.Login(&systemModel.SysUser{
		Username: " spaced.user ",
		Password: "spaced-pass",
	})
	if err != nil {
		t.Fatalf("login by exact whitespace username: %v", err)
	}
	if got.ID != want.ID {
		t.Fatalf("whitespace username login returned ID=%d, want %d", got.ID, want.ID)
	}

	normal := seedLoginUser(t, db, "normal.user", "normal@example.com", "normal-pass")
	got, err = service.Login(&systemModel.SysUser{
		Username: "  normal.user  ",
		Password: "normal-pass",
	})
	if err != nil {
		t.Fatalf("login by normalized username: %v", err)
	}
	if got.ID != normal.ID {
		t.Fatalf("normalized username login returned ID=%d, want %d", got.ID, normal.ID)
	}
}

func TestUserServiceLoginExcludesSoftDeletedEmailUser(t *testing.T) {
	db := setupUserLoginTestDB(t)
	deleted := seedLoginUser(t, db, "deleted.user", "deleted@example.com", "deleted-pass")
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatalf("soft delete login user: %v", err)
	}

	if _, err := new(UserService).Login(&systemModel.SysUser{
		Username: "deleted@example.com",
		Password: "deleted-pass",
	}); err == nil {
		t.Fatal("soft-deleted user logged in by email")
	}
}
