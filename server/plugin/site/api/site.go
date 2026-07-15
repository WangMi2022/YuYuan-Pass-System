package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/site/model"
	siteRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/site/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Site = new(siteAPI)

type siteAPI struct{}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("ID 参数不正确", c)
		return 0, false
	}
	return uint(id), true
}

func (a *siteAPI) Create(c *gin.Context) {
	var site model.SiteBookmark
	if err := c.ShouldBindJSON(&site); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceSite.Create(&site); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(site, "站点创建成功", c)
}

func (a *siteAPI) Update(c *gin.Context) {
	var site model.SiteBookmark
	if err := c.ShouldBindJSON(&site); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceSite.Update(&site); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("站点更新成功", c)
}

func (a *siteAPI) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := serviceSite.Delete(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("站点删除成功", c)
}

func (a *siteAPI) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	site, err := serviceSite.Get(id)
	if err != nil {
		response.FailWithMessage("站点不存在", c)
		return
	}
	response.OkWithData(site, c)
}

func (a *siteAPI) List(c *gin.Context) {
	var search siteRequest.SiteSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceSite.List(search)
	if err != nil {
		global.GVA_LOG.Error("获取站点列表失败", zap.Error(err))
		response.FailWithMessage("获取站点列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}

func (a *siteAPI) Categories(c *gin.Context) {
	list, err := serviceSite.Categories()
	if err != nil {
		response.FailWithMessage("获取站点分类失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *siteAPI) Visit(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	site, err := serviceSite.Visit(id)
	if err != nil {
		response.FailWithMessage("记录访问失败", c)
		return
	}
	response.OkWithData(site, c)
}
