package service

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	assetQueryPrefix = regexp.MustCompile(`^(?:(?:请问|请|帮我|帮忙|查询|查找|搜索|查看|列出|看看|查一下)\s*)+`)
	assetHolderQuery = regexp.MustCompile(`^(.+?)(?:名下|(?:已经|已|都)?(?:认领|领用|领取)(?:了|的)?|(?:保管|负责|持有|使用)的)`)
	assetHolderField = regexp.MustCompile(`^(?:保管人|负责人)(?:是|为|[:：])?\s*(.+?)(?:的)?(?:有哪些|有哪几|资产|设备)`)
	assetHolderName  = regexp.MustCompile(`^[\p{L}\p{N}·_-]+$`)
)

// extractAssetCustodian recognizes a named holder, keeping the department in
// the stored "department-name" form instead of searching the whole sentence.
func extractAssetCustodian(question string) string {
	question = assetQueryPrefix.ReplaceAllString(strings.TrimSpace(question), "")
	match := assetHolderField.FindStringSubmatch(question)
	if len(match) == 0 {
		match = assetHolderQuery.FindStringSubmatch(question)
	}
	if len(match) != 2 {
		return ""
	}
	subject := strings.TrimSpace(match[1])
	for _, suffix := range []string{"已经", "都", "已", "的"} {
		subject = strings.TrimSpace(strings.TrimSuffix(subject, suffix))
	}
	if subject == "" || utf8.RuneCountInString(subject) > 100 ||
		containsAny(subject, "保管人", "负责人", "所有", "全部", "每个", "各部门", "哪些") {
		return ""
	}
	switch subject {
	case "我", "本人", "自己", "大家", "各人", "每人":
		return ""
	}
	subject = strings.NewReplacer("的", "-", "—", "-", "－", "-", "/", "-").Replace(subject)
	subject = strings.Join(strings.Fields(subject), "-")
	if !assetHolderName.MatchString(subject) {
		return ""
	}
	return subject
}
