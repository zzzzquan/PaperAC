package parser

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

// 句末标点符号正则
var sentenceEndPunctuationRe = regexp.MustCompile(`[。？！；.?!;]$`)

// 标题模式正则（中文章节标题）
var chapterTitleRe = regexp.MustCompile(`^(第[一二三四五六七八九十百千\d]+[章节部分篇]|摘\s*要|Abstract|引\s*言|结\s*论|参考文献|致\s*谢|附\s*录)`)

// ExtractText 读取PDF文件并提取纯文本，自动合并软换行
func ExtractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", err
	}

	// 对提取的文本进行后处理，合并软换行
	normalized := normalizeText(buf.String())

	// 移除所有空格（包括中文空格和英文空格）
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "　", "") // 全角空格

	return normalized, nil
}

// normalizeText 合并PDF中的软换行
// 规则：
// 1. 如果一行不以句末标点结尾，且下一行不是空行或标题，则合并
// 2. 保留真正的段落分隔（空行或以标点结尾的行）
// 3. 识别并保留标题行
func normalizeText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return ""
	}

	var result []string
	var currentParagraph strings.Builder

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 空行表示段落分隔
		if trimmed == "" {
			if currentParagraph.Len() > 0 {
				result = append(result, currentParagraph.String())
				currentParagraph.Reset()
			}
			continue
		}

		// 检查是否是标题行
		if isLikelyTitle(trimmed) {
			// 先保存当前段落
			if currentParagraph.Len() > 0 {
				result = append(result, currentParagraph.String())
				currentParagraph.Reset()
			}
			// 标题作为独立行
			result = append(result, trimmed)
			continue
		}

		// 判断前一行是否以句末标点结尾
		if currentParagraph.Len() > 0 {
			prevContent := currentParagraph.String()
			if sentenceEndPunctuationRe.MatchString(prevContent) {
				// 前一行以标点结尾，当前行是新段落
				result = append(result, prevContent)
				currentParagraph.Reset()
				currentParagraph.WriteString(trimmed)
			} else {
				// 前一行不以标点结尾，合并（不添加空格，因为中文不需要）
				// 但如果是英文/数字结尾且下一行是英文/数字开头，需要添加空格
				if needsSpaceBetween(prevContent, trimmed) {
					currentParagraph.WriteString(" ")
				}
				currentParagraph.WriteString(trimmed)
			}
		} else {
			currentParagraph.WriteString(trimmed)
		}

		// 如果是最后一行或当前行以标点结尾
		if i == len(lines)-1 {
			if currentParagraph.Len() > 0 {
				result = append(result, currentParagraph.String())
			}
		}
	}

	return strings.Join(result, "\n")
}

// isLikelyTitle 判断是否可能是标题
// 注意：在PDF解析阶段，只识别非常明确的章节标题，避免误将普通短行当作标题
func isLikelyTitle(line string) bool {
	// 太长的行不太可能是标题
	runeCount := len([]rune(line))
	if runeCount > 50 {
		return false
	}

	// 只匹配明确的章节标题模式
	return chapterTitleRe.MatchString(line)
}

// needsSpaceBetween 判断两段文本合并时是否需要添加空格
func needsSpaceBetween(prev, next string) bool {
	if len(prev) == 0 || len(next) == 0 {
		return false
	}

	prevRunes := []rune(prev)
	nextRunes := []rune(next)

	lastChar := prevRunes[len(prevRunes)-1]
	firstChar := nextRunes[0]

	// 如果前一个字符是英文/数字，且后一个也是英文/数字，需要空格
	prevIsLatin := unicode.IsLetter(lastChar) && lastChar < 256 || unicode.IsDigit(lastChar)
	nextIsLatin := unicode.IsLetter(firstChar) && firstChar < 256 || unicode.IsDigit(firstChar)

	return prevIsLatin && nextIsLatin
}
