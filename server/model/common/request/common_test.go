package request

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type paginationTestRow struct {
	ID uint
}

func (paginationTestRow) TableName() string { return "pagination_test_rows" }

func TestPaginateNormalizesAndBoundsQueries(t *testing.T) {
	tests := []struct {
		name       string
		pageInfo   PageInfo
		max        int
		wantPage   int
		wantSize   int
		wantOffset int
	}{
		{name: "defaults invalid input", pageInfo: PageInfo{Page: -2, PageSize: 0}, wantPage: 1, wantSize: 10},
		{name: "caps regular collections", pageInfo: PageInfo{Page: 3, PageSize: 250}, wantPage: 3, wantSize: 100, wantOffset: 200},
		{name: "supports explicit bounded bulk selection", pageInfo: PageInfo{Page: 2, PageSize: 999}, max: 1000, wantPage: 2, wantSize: 999, wantOffset: 999},
		{name: "invalid max uses shared cap", pageInfo: PageInfo{Page: 2, PageSize: 500}, max: -1, wantPage: 2, wantSize: 100, wantOffset: 100},
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageInfo := tt.pageInfo
			var scope func(*gorm.DB) *gorm.DB
			if tt.max == 0 {
				scope = pageInfo.Paginate()
			} else {
				scope = pageInfo.PaginateWithMax(tt.max)
			}
			statement := db.Session(&gorm.Session{DryRun: true}).Scopes(scope).Find(&[]paginationTestRow{}).Statement
			limitClause, ok := statement.Clauses["LIMIT"].Expression.(clause.Limit)
			if !ok || limitClause.Limit == nil {
				t.Fatalf("pagination scope did not create a LIMIT clause: %#v", statement.Clauses["LIMIT"])
			}
			if pageInfo.Page != tt.wantPage || pageInfo.PageSize != tt.wantSize {
				t.Fatalf("normalized page = (%d, %d), want (%d, %d)", pageInfo.Page, pageInfo.PageSize, tt.wantPage, tt.wantSize)
			}
			if *limitClause.Limit != tt.wantSize || limitClause.Offset != tt.wantOffset {
				t.Fatalf("query pagination = limit %d offset %d, want limit %d offset %d", *limitClause.Limit, limitClause.Offset, tt.wantSize, tt.wantOffset)
			}
		})
	}
}
