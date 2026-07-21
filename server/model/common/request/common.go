package request

import (
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     int    `json:"page" form:"page"`         // 页码
	PageSize int    `json:"pageSize" form:"pageSize"` // 每页大小
	Keyword  string `json:"keyword" form:"keyword"`   // 关键字
}

func (r *PageInfo) Paginate() func(db *gorm.DB) *gorm.DB {
	return r.PaginateWithMax(MaxPageSize)
}

// PaginateWithMax applies the shared pagination rules with an explicit upper
// bound for the few collection endpoints that intentionally support bulk
// selection. A non-positive maxPageSize falls back to MaxPageSize.
func (r *PageInfo) PaginateWithMax(maxPageSize int) func(db *gorm.DB) *gorm.DB {
	offset, limit := r.normalize(maxPageSize)
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}

func (r *PageInfo) normalize(maxPageSize int) (offset, limit int) {
	if maxPageSize <= 0 {
		maxPageSize = MaxPageSize
	}
	if r.Page <= 0 {
		r.Page = DefaultPage
	}
	switch {
	case r.PageSize > maxPageSize:
		r.PageSize = maxPageSize
	case r.PageSize <= 0:
		r.PageSize = DefaultPageSize
	}

	maxInt := int(^uint(0) >> 1)
	if r.Page-1 > maxInt/r.PageSize {
		return maxInt, r.PageSize
	}
	return (r.Page - 1) * r.PageSize, r.PageSize
}

// GetById Find by id structure
type GetById struct {
	ID int `json:"id" form:"id"` // 主键ID
}

func (r *GetById) Uint() uint {
	return uint(r.ID)
}

type IdsReq struct {
	Ids []int `json:"ids" form:"ids"`
}

// GetAuthorityId Get role by id structure
type GetAuthorityId struct {
	AuthorityId uint `json:"authorityId" form:"authorityId"` // 角色ID
}

type Empty struct{}
