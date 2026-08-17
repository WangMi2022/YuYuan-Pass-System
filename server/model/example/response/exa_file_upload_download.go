package response

import "github.com/WangMi2022/mit-assets-admin/server/model/example"

type ExaFileResponse struct {
	File example.ExaFileUploadAndDownload `json:"file"`
}
