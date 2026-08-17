package api

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	commonResponse "github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/service"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type smartAPI struct{}

var SmartAPI = new(smartAPI)

type copilotQueryRequest struct {
	Question  string `json:"question" binding:"required"`
	SessionID uint   `json:"sessionId"`
}

type reportSearchRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type announcementDraftRequest struct {
	AnnouncementID uint `json:"announcementId" binding:"required"`
}

type draftListRequest struct {
	DraftType string `form:"draftType"`
}

type operationAssetCandidatesRequest struct {
	OperationType string `form:"operationType"`
	Keyword       string `form:"keyword"`
}

func parseID(c *gin.Context, key string) (uint, error) {
	raw := strings.TrimSpace(c.Query(key))
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("ID 参数不正确")
	}
	return uint(id), nil
}

func (a *smartAPI) Query(c *gin.Context) {
	var req copilotQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResponse.FailWithMessage("问题参数不正确", c)
		return
	}
	result, err := Smart.Query(c.Request.Context(), utils.GetUserID(c), utils.GetUserAuthorityId(c), req.Question, req.SessionID)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithData(result, c)
}

// QueryStream keeps the SSE contract for clients that want progressive UI.
// The first release emits one audited, complete event so no partial model text
// can be mistaken for a completed business result.
func (a *smartAPI) QueryStream(c *gin.Context) {
	var req copilotQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Header("Content-Type", "text/event-stream")
		_, _ = c.Writer.WriteString("data: {\"code\":7,\"msg\":\"问题参数不正确\"}\n\n")
		return
	}
	result, err := Smart.Query(c.Request.Context(), utils.GetUserID(c), utils.GetUserAuthorityId(c), req.Question, req.SessionID)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	payload := gin.H{"code": 0, "data": result, "msg": "成功"}
	if err != nil {
		payload = gin.H{"code": 7, "data": map[string]any{}, "msg": err.Error()}
	}
	encoded, _ := json.Marshal(payload)
	_, _ = c.Writer.WriteString("data: " + string(encoded) + "\n\n")
	c.Writer.Flush()
}

func (a *smartAPI) Sessions(c *gin.Context) {
	list, err := Smart.Sessions(utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		commonResponse.FailWithMessage("获取会话失败", c)
		return
	}
	commonResponse.OkWithData(list, c)
}

func (a *smartAPI) Session(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	session, messages, err := Smart.Session(utils.GetUserID(c), utils.GetUserAuthorityId(c), id)
	if err != nil {
		commonResponse.FailWithMessage("会话不存在或无权访问", c)
		return
	}
	commonResponse.OkWithData(gin.H{"session": session, "messages": messages}, c)
}

func (a *smartAPI) DeleteSession(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	if err = Smart.DeleteSession(utils.GetUserID(c), utils.GetUserAuthorityId(c), id); err != nil {
		commonResponse.FailWithMessage("删除会话失败", c)
		return
	}
	commonResponse.OkWithMessage("会话已删除", c)
}

func (a *smartAPI) Tools(c *gin.Context) {
	commonResponse.OkWithData(Smart.Tools(utils.GetUserAuthorityId(c)), c)
}

func (a *smartAPI) TodayReport(c *gin.Context) {
	report, err := Smart.TodayReport(c.Request.Context(), utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		commonResponse.FailWithMessage("生成今日智能日报失败: "+err.Error(), c)
		return
	}
	commonResponse.OkWithData(report, c)
}

func (a *smartAPI) Reports(c *gin.Context) {
	var req reportSearchRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := Smart.Reports(utils.GetUserID(c), utils.GetUserAuthorityId(c), service.ReportListInput{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		commonResponse.FailWithMessage("获取日报列表失败", c)
		return
	}
	commonResponse.OkWithDetailed(commonResponse.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取成功", c)
}

func (a *smartAPI) Report(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	report, err := Smart.Report(utils.GetUserID(c), utils.GetUserAuthorityId(c), id)
	if err != nil {
		commonResponse.FailWithMessage("日报不存在或无权访问", c)
		return
	}
	commonResponse.OkWithData(report, c)
}

func (a *smartAPI) GenerateReport(c *gin.Context) {
	report, err := Smart.GenerateReport(c.Request.Context(), utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		commonResponse.FailWithMessage("生成日报失败: "+err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(report, "日报已生成", c)
}

func (a *smartAPI) Subscription(c *gin.Context) {
	item, err := Smart.Subscription(utils.GetUserID(c))
	if err != nil {
		commonResponse.FailWithMessage("获取日报订阅失败", c)
		return
	}
	commonResponse.OkWithData(item, c)
}

func (a *smartAPI) SaveSubscription(c *gin.Context) {
	var input model.SmartReportSubscription
	if err := c.ShouldBindJSON(&input); err != nil {
		commonResponse.FailWithMessage("订阅参数不正确", c)
		return
	}
	item, err := Smart.SaveSubscription(utils.GetUserID(c), input)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(item, "日报订阅已保存", c)
}

func (a *smartAPI) Deliveries(c *gin.Context) {
	list, err := Smart.Deliveries(utils.GetUserID(c), 30)
	if err != nil {
		commonResponse.FailWithMessage("获取日报发送记录失败", c)
		return
	}
	commonResponse.OkWithData(list, c)
}

func (a *smartAPI) ExtractAnnouncement(c *gin.Context) {
	var req announcementDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResponse.FailWithMessage("公告参数不正确", c)
		return
	}
	draft, err := Smart.AnnouncementDraft(utils.GetUserID(c), req.AnnouncementID)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(draft, "公告已提取为日程草稿", c)
}

func (a *smartAPI) OperationDraft(c *gin.Context) {
	var input service.OperationDraftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		commonResponse.FailWithMessage("业务草稿参数不正确", c)
		return
	}
	draft, err := Smart.OperationDraft(utils.GetUserID(c), utils.GetUserAuthorityId(c), input)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(draft, "业务单草稿已生成，请确认后提交", c)
}

func (a *smartAPI) OperationAssetCandidates(c *gin.Context) {
	var req operationAssetCandidatesRequest
	_ = c.ShouldBindQuery(&req)
	list, err := Smart.OperationAssetCandidates(utils.GetUserAuthorityId(c), service.OperationAssetCandidateInput{OperationType: req.OperationType, Keyword: req.Keyword})
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithData(list, c)
}

func (a *smartAPI) Drafts(c *gin.Context) {
	var req draftListRequest
	_ = c.ShouldBindQuery(&req)
	list, err := Smart.Drafts(utils.GetUserID(c), req.DraftType)
	if err != nil {
		commonResponse.FailWithMessage("获取智能草稿失败", c)
		return
	}
	commonResponse.OkWithData(list, c)
}

func (a *smartAPI) AcceptDraft(c *gin.Context) {
	var req service.DraftAcceptInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commonResponse.FailWithMessage("草稿参数不正确", c)
		return
	}
	result, err := Smart.AcceptDraft(c.Request.Context(), utils.GetUserID(c), utils.GetUserAuthorityId(c), utils.GetUserName(c), req.ID)
	if err != nil {
		commonResponse.FailWithMessage(err.Error(), c)
		return
	}
	commonResponse.OkWithDetailed(result, "草稿已确认并创建业务记录", c)
}
