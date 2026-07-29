package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	commonResponse "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	"github.com/gin-gonic/gin"
)

func TestInvoiceSearchBindingTreatsEmptyDatesAsUnset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(http.MethodGet, "/invoice/list?page=1&pageSize=20&keyword=&status=&direction=&startDate=&endDate=", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request

	var search invoiceRequest.InvoiceSearch
	if err := context.ShouldBindQuery(&search); err != nil {
		t.Fatalf("empty optional dates should bind successfully: %v", err)
	}
	search.Normalize()
	if search.StartDate != nil || search.EndDate != nil {
		t.Fatalf("empty optional dates should remain unset: start=%v end=%v", search.StartDate, search.EndDate)
	}
}

func TestUploadRejectsMoreThanFiveFilesPerBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for index := 0; index < maxInvoiceUploadBatch+1; index++ {
		part, err := writer.CreateFormFile("files", fmt.Sprintf("invoice-%d.png", index))
		if err != nil {
			t.Fatalf("create multipart file: %v", err)
		}
		if _, err = part.Write([]byte("not inspected because the batch is rejected first")); err != nil {
			t.Fatalf("write multipart file: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/invoice/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	(invoiceAPI{}).Upload(context)

	var response commonResponse.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if response.Code != commonResponse.ERROR || response.Msg != "每批最多上传 5 张发票图片" {
		t.Fatalf("unexpected upload response: %#v", response)
	}
}
