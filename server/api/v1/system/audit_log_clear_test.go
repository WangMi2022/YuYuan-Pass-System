package system

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/gin-gonic/gin"
)

func TestResolveLogClearRangeRequiresUnambiguousScope(t *testing.T) {
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name        string
		request     request.LogDeleteReq
		shouldClear bool
		wantError   bool
	}{
		{name: "all logs", request: request.LogDeleteReq{ClearAll: true}, shouldClear: true},
		{name: "older than end", request: request.LogDeleteReq{LogTimeRange: request.LogTimeRange{EndTime: &earlier}}, shouldClear: true},
		{name: "custom range", request: request.LogDeleteReq{LogTimeRange: request.LogTimeRange{StartTime: &earlier, EndTime: &now}}, shouldClear: true},
		{name: "batch ids", request: request.LogDeleteReq{Ids: []int{1}}, shouldClear: false},
		{name: "all plus range", request: request.LogDeleteReq{ClearAll: true, LogTimeRange: request.LogTimeRange{EndTime: &now}}, wantError: true},
		{name: "reversed range", request: request.LogDeleteReq{LogTimeRange: request.LogTimeRange{StartTime: &now, EndTime: &earlier}}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, shouldClear, err := resolveLogClearRange(test.request)
			if (err != nil) != test.wantError {
				t.Fatalf("resolve error = %v, wantError %v", err, test.wantError)
			}
			if shouldClear != test.shouldClear {
				t.Fatalf("shouldClear = %v, want %v", shouldClear, test.shouldClear)
			}
		})
	}
}

func TestLogDeleteReqBindsISOTimeRangeFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"DELETE",
		"/logs?startTime=2026-07-01T00%3A00%3A00.000Z&endTime=2026-07-15T23%3A59%3A59.000Z",
		nil,
	)

	var bound request.LogDeleteReq
	if err := context.ShouldBindQuery(&bound); err != nil {
		t.Fatalf("bind log clear range: %v", err)
	}
	if bound.StartTime == nil || bound.EndTime == nil {
		t.Fatalf("bound range is incomplete: %#v", bound)
	}
	if bound.StartTime.Format(time.RFC3339Nano) != "2026-07-01T00:00:00Z" ||
		bound.EndTime.Format(time.RFC3339Nano) != "2026-07-15T23:59:59Z" {
		t.Fatalf("unexpected bound range: %s - %s", bound.StartTime, bound.EndTime)
	}
}
