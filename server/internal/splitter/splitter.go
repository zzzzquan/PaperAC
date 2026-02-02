package splitter

import (
	"regexp"
	"strings"
	"unicode"
)

// SegmentType 表示分段类型
type SegmentType string

const (
	SegmentBody    SegmentType = "body"    // 正文（参与AIGC检测）
	SegmentTitle   SegmentType = "title"   // 标题
	SegmentTable   SegmentType = "table"   // 表格行
	SegmentList    SegmentType = "list"    // 列表项
	SegmentMeta    SegmentType = "meta"    // 元信息（封面、表头等，不参与检测）
	SegmentNewline SegmentType = "newline" // 硬换行
)

// Segment 表示一个带类型的文本段落
type Segment struct {
	Text string      `json:"text"`
	Type SegmentType `json:"type"`
}

// 句末标点正则
var sentenceEndRe = regexp.MustCompile(`([。？！；!?;]+)`)

// 章节标题正则
var titlePatternRe = regexp.MustCompile(`^(第[一二三四五六七八九十百千\d]+[章节部分篇]|摘\s*要|Abstract|引\s*言|结\s*论|参考文献|致\s*谢|附\s*录|关键词|Keywords)`)

// 列表项正则
var listPatternRe = regexp.MustCompile(`^(\d+[.、)）]|\([0-9a-zA-Z]+\)|[•●○■□★☆\-–—]|[a-zA-Z][.、)）])`)

// 表格行正则（包含多个连续空格或制表符分隔的内容）
var tablePatternRe = regexp.MustCompile(`\s{2,}|\t`)

// 特殊字符正则（用于过滤不适合检测的句子）
var specialCharRe = regexp.MustCompile(`[#$%&*+/<=>@\\^_` + "`" + `{|}~\[\]\(\)\-×÷∑∏∫∂√∞≈≠≤≥±°′″αβγδεζηθικλμνξπρστυφχψω·]`)

// 论文封面/表头信息正则
var coverInfoRe = regexp.MustCompile(`(学号|姓名|班级|专业|学院|书院|题目|课程|论文|评阅|成绩|指导|教师|序号|负责|组别|全员|全英文|专题[一二三四五六七八九十]|\d{10,})`)

// 纯数字或学号模式
var pureNumberRe = regexp.MustCompile(`^\d+$|1[12]\d{8,}`)

// Split 将长文本切分为句子（保持向后兼容）
func Split(text string) []string {
	segments := SplitWithStructure(text)
	var result []string
	for _, seg := range segments {
		if seg.Type != SegmentNewline {
			result = append(result, seg.Text)
		}
	}
	return result
}

// SplitWithStructure 将文本切分为带类型的分段
func SplitWithStructure(text string) []Segment {
	var segments []Segment

	// 按换行符分割，保留原本的行结构
	// 使用 strings.Split 会把换行符吃掉，我们在循环末尾补上 SegmentNewline
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		// 移除首尾空白，但保留内容
		trimmed := strings.TrimSpace(line)

		if trimmed != "" {
			segType := detectSegmentType(trimmed)
			if segType == SegmentBody {
				// 正文分句
				sentences := splitSentences(trimmed)
				for _, s := range sentences {
					s = strings.TrimSpace(s)
					if s == "" {
						continue
					}
					if shouldFilter(s) {
						segments = append(segments, Segment{Text: s, Type: SegmentMeta})
					} else {
						segments = append(segments, Segment{Text: s, Type: SegmentBody})
					}
				}
			} else {
				// 非正文保持原样
				segments = append(segments, Segment{Text: trimmed, Type: segType})
			}
		}

		// 除非是最后一行，否则每一行结束后都插入一个硬换行
		if i < len(lines)-1 {
			segments = append(segments, Segment{Text: "", Type: SegmentNewline})
		}
	}

	return segments
}

// shouldFilter 检查句子是否应该被过滤掉
func shouldFilter(s string) bool {
	if specialCharRe.MatchString(s) {
		return true
	}
	if coverInfoRe.MatchString(s) {
		return true
	}
	if pureNumberRe.MatchString(s) {
		return true
	}
	return false
}

// detectSegmentType 检测段落类型
func detectSegmentType(para string) SegmentType {
	if listPatternRe.MatchString(para) {
		return SegmentList
	}
	if isTableRow(para) {
		return SegmentTable
	}
	if isTitle(para) {
		return SegmentTitle
	}
	return SegmentBody
}

// isTitle 判断是否是标题
func isTitle(para string) bool {
	runeCount := len([]rune(para))
	if runeCount > 50 {
		return false
	}
	if titlePatternRe.MatchString(para) {
		return true
	}
	if strings.HasSuffix(para, "：") || strings.HasSuffix(para, ":") {
		return false
	}
	if runeCount <= 25 && !sentenceEndRe.MatchString(para) {
		connectors := []string{"的", "和", "与", "或", "但", "而", "是", "在", "了", "有", "为", "对", "从", "到"}
		for _, c := range connectors {
			if strings.HasSuffix(para, c) {
				return false
			}
		}
		chineseCount := 0
		for _, r := range para {
			if unicode.Is(unicode.Han, r) {
				chineseCount++
			}
		}
		if float64(chineseCount)/float64(runeCount) > 0.5 {
			return true
		}
	}
	return false
}

// isTableRow 判断是否是表格行
func isTableRow(para string) bool {
	matches := tablePatternRe.FindAllStringIndex(para, -1)
	if len(matches) >= 2 {
		parts := tablePatternRe.Split(para, -1)
		shortParts := 0
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" && len([]rune(p)) <= 20 {
				shortParts++
			}
		}
		return shortParts >= 3
	}
	return false
}

// splitSentences 将段落分割成句子
func splitSentences(para string) []string {
	temp := sentenceEndRe.ReplaceAllString(para, "$1\n")
	lines := strings.Split(temp, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}
