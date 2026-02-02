package splitter

import (
	"regexp"
	"strings"
	"unicode"
)

// SegmentType 表示分段类型
type SegmentType string

const (
	SegmentBody  SegmentType = "body"  // 正文
	SegmentTitle SegmentType = "title" // 标题
	SegmentTable SegmentType = "table" // 表格行
	SegmentList  SegmentType = "list"  // 列表项
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
// 包括：数学符号、公式符号、特殊符号、网址、代码相关符号、减号、间隔号等
var specialCharRe = regexp.MustCompile(`[#$%&*+/<=>@\\^_` + "`" + `{|}~\[\]（）\(\)\-×÷∑∏∫∂√∞≈≠≤≥±°′″αβγδεζηθικλμνξπρστυφχψω·]`)

// Split 将长文本切分为句子（保持向后兼容）
func Split(text string) []string {
	segments := SplitWithStructure(text)
	var result []string
	for _, seg := range segments {
		result = append(result, seg.Text)
	}
	return result
}

// SplitWithStructure 将文本切分为带类型的分段
func SplitWithStructure(text string) []Segment {
	var segments []Segment

	// 按空行或换行分割成段落
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		segType := detectSegmentType(para)

		if segType == SegmentBody {
			// 对正文进行分句
			sentences := splitSentences(para)
			for _, s := range sentences {
				s = strings.TrimSpace(s)
				if s != "" && !hasSpecialCharacters(s) {
					// 只保留不含特殊字符的句子
					segments = append(segments, Segment{Text: s, Type: SegmentBody})
				}
			}
		} else {
			// 非正文类型保持原样（但也过滤特殊字符）
			if !hasSpecialCharacters(para) {
				segments = append(segments, Segment{Text: para, Type: segType})
			}
		}
	}

	return segments
}

// hasSpecialCharacters 检查句子是否含有特殊字符
// 返回 true 表示含有特殊字符，应该被过滤掉
func hasSpecialCharacters(s string) bool {
	return specialCharRe.MatchString(s)
}

// detectSegmentType 检测段落类型
func detectSegmentType(para string) SegmentType {
	// 优先检查列表项（有明确的格式特征）
	if listPatternRe.MatchString(para) {
		return SegmentList
	}

	// 检查是否是表格行（包含多个连续空格或制表符）
	if isTableRow(para) {
		return SegmentTable
	}

	// 检查是否是标题
	if isTitle(para) {
		return SegmentTitle
	}

	return SegmentBody
}

// isTitle 判断是否是标题
func isTitle(para string) bool {
	runeCount := len([]rune(para))

	// 太长不可能是标题
	if runeCount > 50 {
		return false
	}

	// 匹配章节标题模式
	if titlePatternRe.MatchString(para) {
		return true
	}

	// 以冒号结尾的通常是引导语，不是标题
	if strings.HasSuffix(para, "：") || strings.HasSuffix(para, ":") {
		return false
	}

	// 短文本且不以句末标点结尾
	if runeCount <= 25 && !sentenceEndRe.MatchString(para) {
		// 排除以连接词结尾的句子片段
		connectors := []string{"的", "和", "与", "或", "但", "而", "是", "在", "了", "有", "为", "对", "从", "到"}
		for _, c := range connectors {
			if strings.HasSuffix(para, c) {
				return false
			}
		}
		// 检查是否主要是中文字符（标题通常是中文）
		chineseCount := 0
		for _, r := range para {
			if unicode.Is(unicode.Han, r) {
				chineseCount++
			}
		}
		// 如果超过一半是中文字符，且较短，认为是标题
		if float64(chineseCount)/float64(runeCount) > 0.5 {
			return true
		}
	}

	return false
}

// isTableRow 判断是否是表格行
func isTableRow(para string) bool {
	// 表格行特征：包含多个连续空格或制表符分隔的内容
	matches := tablePatternRe.FindAllStringIndex(para, -1)
	// 如果有多个分隔符，可能是表格行
	if len(matches) >= 2 {
		// 进一步验证：检查是否有数字或短文本片段
		parts := tablePatternRe.Split(para, -1)
		shortParts := 0
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" && len([]rune(p)) <= 20 {
				shortParts++
			}
		}
		// 如果大部分是短片段，认为是表格行
		return shortParts >= 3
	}
	return false
}

// splitSentences 将段落分割成句子
func splitSentences(para string) []string {
	// 使用标点符号分割，保留标点
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
