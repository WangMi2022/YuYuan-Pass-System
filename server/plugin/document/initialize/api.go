package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

var apiRules = []system.SysApi{
	{Path: "/document/upload", Description: "上传文档", ApiGroup: "文档管理", Method: "POST"},
	{Path: "/document/list", Description: "文档列表", ApiGroup: "文档管理", Method: "GET"},
	{Path: "/document/detail", Description: "文档详情", ApiGroup: "文档管理", Method: "GET"},
	{Path: "/document/file", Description: "读取文档原文件", ApiGroup: "文档管理", Method: "GET"},
	{Path: "/document/updateContent", Description: "保存在线文档", ApiGroup: "文档管理", Method: "PUT"},
	{Path: "/document/delete", Description: "删除文档", ApiGroup: "文档管理", Method: "DELETE"},
}

func Api(_ context.Context) { utils.RegisterApis(apiRules...) }
