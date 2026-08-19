package model

import (
	"crypto/subtle"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AssetStatusPendingInbound = "pending_inbound"
	AssetStatusIdle           = "idle"
	AssetStatusInUse          = "in_use"
	AssetStatusMaintenance    = "maintenance"
	AssetStatusRetired        = "retired"
)

// Photo 保存资产图片在 RustFS/MinIO 中的对象信息。
type Photo struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	AssetID     uint   `json:"assetId,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
}

func photoObjectPrefix() string {
	prefix := strings.Trim(strings.TrimSpace(global.GVA_CONFIG.Minio.BasePath), "/")
	if prefix == "" {
		return "uploads"
	}
	return prefix
}

// ValidPhotoObjectKey 只允许访问当前资产图片目录中的规范对象路径。
func ValidPhotoObjectKey(key string) bool {
	if key == "" || key != strings.TrimSpace(key) || strings.Contains(key, "..") || strings.ContainsAny(key, "\\\x00\r\n") {
		return false
	}
	parts := strings.Split(key, "/")
	if !strings.HasPrefix(key, photoObjectPrefix()+"/") {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

// managedPhotoObjectKeyFromURL 仅从当前配置的 MinIO Bucket 直链中恢复旧记录的对象 key。
func managedPhotoObjectKeyFromURL(rawURL string) (string, bool) {
	bucketURL, err := url.Parse(strings.TrimSpace(global.GVA_CONFIG.Minio.BucketUrl))
	if err != nil || bucketURL.Scheme == "" || bucketURL.Host == "" || bucketURL.User != nil || bucketURL.RawQuery != "" || bucketURL.Fragment != "" {
		return "", false
	}
	candidate, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || candidate.Scheme == "" || candidate.Host == "" || candidate.User != nil || candidate.RawQuery != "" || candidate.Fragment != "" {
		return "", false
	}
	if !strings.EqualFold(candidate.Scheme, bucketURL.Scheme) || !strings.EqualFold(candidate.Host, bucketURL.Host) {
		return "", false
	}
	bucketPath := strings.TrimRight(bucketURL.Path, "/")
	objectPrefix := bucketPath + "/"
	if bucketPath == "" || !strings.HasPrefix(candidate.Path, objectPrefix) {
		return "", false
	}
	key := strings.TrimPrefix(candidate.Path, objectPrefix)
	if !ValidPhotoObjectKey(key) {
		return "", false
	}
	return key, true
}

// ResolvePhotoObjectKey 兼容新记录的 key 与旧记录中可信的 MinIO 直链。
func ResolvePhotoObjectKey(photo Photo) (string, bool) {
	key := strings.TrimSpace(photo.Key)
	if ValidPhotoObjectKey(key) {
		return key, true
	}
	return managedPhotoObjectKeyFromURL(photo.URL)
}

// NormalizeAssetPhoto 将旧图片记录在内存中规范化为受鉴权的后端代理地址。
func NormalizeAssetPhoto(photo *Photo, assetID uint) {
	if photo == nil {
		return
	}
	if assetID > 0 {
		photo.AssetID = assetID
		photo.AccessToken = ""
	}
	key, ok := ResolvePhotoObjectKey(*photo)
	if !ok {
		return
	}
	photo.Key = key
	photo.URL = BuildPhotoURL(photo.AssetID, key, photo.AccessToken)
}

type photoTokenClaims struct {
	UserID  uint   `json:"uid"`
	Key     string `json:"key"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

func CreatePhotoAccessToken(userID uint, key string) (string, error) {
	secret := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	key = strings.TrimSpace(key)
	if userID == 0 || key == "" || secret == "" {
		return "", errors.New("图片访问令牌参数不完整")
	}
	now := time.Now()
	claims := photoTokenClaims{
		UserID:  userID,
		Key:     key,
		Purpose: "asset-photo",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "asset-photo",
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ValidatePhotoAccessToken(tokenString string, userID uint, key string) bool {
	secret := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	if secret == "" || tokenString == "" || userID == 0 || key == "" {
		return false
	}
	claims := &photoTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("图片访问令牌算法不正确")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return false
	}
	return claims.UserID == userID && subtle.ConstantTimeCompare([]byte(claims.Key), []byte(key)) == 1 && claims.Purpose == "asset-photo"
}

func BuildPhotoURL(assetID uint, key, accessToken string) string {
	query := "?key=" + url.QueryEscape(key)
	if assetID > 0 {
		query += "&assetId=" + strconv.FormatUint(uint64(assetID), 10)
	}
	if accessToken != "" {
		query += "&token=" + url.QueryEscape(accessToken)
	}
	return "/api/asset/photo" + query
}

// Category 资产分类。
type Category struct {
	global.GVA_MODEL
	Name        string `json:"name" form:"name" gorm:"size:100;not null;uniqueIndex;comment:分类名称"`
	Code        string `json:"code" form:"code" gorm:"size:50;not null;uniqueIndex;comment:分类编码"`
	Description string `json:"description" form:"description" gorm:"size:500;comment:分类说明"`
	Color       string `json:"color" form:"color" gorm:"size:20;default:#334155;comment:展示颜色"`
	Sort        int    `json:"sort" form:"sort" gorm:"default:0;comment:排序"`
	Enabled     bool   `json:"enabled" form:"enabled" gorm:"default:true;comment:是否启用"`
}

func (Category) TableName() string { return "asset_categories" }

// Asset 资产档案，支持按数量和单价计算原值，并独立记录当前估值。
type Asset struct {
	global.GVA_MODEL
	AssetCode       string     `json:"assetCode" form:"assetCode" gorm:"size:80;not null;uniqueIndex;comment:资产编号"`
	Name            string     `json:"name" form:"name" gorm:"size:150;not null;index;comment:资产名称"`
	CategoryID      uint       `json:"categoryId" form:"categoryId" gorm:"not null;index;index:idx_assets_category_status,priority:1;comment:资产分类ID"`
	Category        Category   `json:"category" gorm:"foreignKey:CategoryID"`
	Brand           string     `json:"brand" form:"brand" gorm:"size:100;comment:品牌"`
	Model           string     `json:"model" form:"model" gorm:"size:120;comment:规格型号"`
	SerialNumber    string     `json:"serialNumber" form:"serialNumber" gorm:"size:120;index;comment:序列号"`
	Specifications  string     `json:"specifications" form:"specifications" gorm:"size:1000;comment:规格参数"`
	ProductionDate  *time.Time `json:"productionDate" form:"productionDate" gorm:"type:date;comment:生产日期"`
	Quantity        int        `json:"quantity" form:"quantity" gorm:"not null;default:1;comment:数量"`
	Unit            string     `json:"unit" form:"unit" gorm:"size:30;default:件;comment:计量单位"`
	UnitPrice       float64    `json:"unitPrice" form:"unitPrice" gorm:"type:numeric(16,2);not null;default:0;comment:采购单价"`
	OriginalValue   float64    `json:"originalValue" gorm:"type:numeric(18,2);not null;default:0;comment:资产原值"`
	CurrentValue    float64    `json:"currentValue" form:"currentValue" gorm:"type:numeric(18,2);not null;default:0;comment:当前估值"`
	Status          string     `json:"status" form:"status" gorm:"size:30;not null;default:pending_inbound;index;index:idx_assets_category_status,priority:2;comment:资产状态"`
	Location        string     `json:"location" form:"location" gorm:"size:150;index;comment:存放位置"`
	Custodian       string     `json:"custodian" form:"custodian" gorm:"size:100;index;comment:保管人"`
	Supplier        string     `json:"supplier" form:"supplier" gorm:"size:150;comment:供应商"`
	PurchaseDate    *time.Time `json:"purchaseDate" form:"purchaseDate" gorm:"type:date;comment:购置日期"`
	WarrantyEndDate *time.Time `json:"warrantyEndDate" form:"warrantyEndDate" gorm:"type:date;comment:质保到期日"`
	Photos          []Photo    `json:"photos" gorm:"serializer:json;type:jsonb;comment:资产图片"`
	Remarks         string     `json:"remarks" form:"remarks" gorm:"type:text;comment:备注"`
}

// NormalizeAssetPhotos 在内存中补齐旧资产图片的 key、资产归属与代理 URL。
func NormalizeAssetPhotos(asset *Asset) {
	if asset == nil {
		return
	}
	for index := range asset.Photos {
		NormalizeAssetPhoto(&asset.Photos[index], asset.ID)
	}
}

func (Asset) TableName() string { return "assets" }
