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

// 纯数字正则
var numericLineRe = regexp.MustCompile(`^\d+$`)

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

	rawText := buf.String()

	// 预处理：
	// 1. 合并被拆散的标题/短语（跨行或跨空行的单字）
	// 2. 合并纯数字行（加空格）
	mergedText := preprocessLines(rawText)

	// 对提取的文本进行后处理，合并软换行
	normalized := normalizeText(mergedText)

	// 注意：之前这里移除了所有空格，导致我们添加的空格也被移除了
	// 现在移除这个操作。normalizeText 产生的段落应该已经处理好了空格问题（通过 needsSpaceBetween）
	// normalized = strings.ReplaceAll(normalized, " ", "")

	// 但我们可能仍然想要移除全角空格，或者多余的空格？
	// 暂时只移除全角空格
	normalized = strings.ReplaceAll(normalized, "　", "") // 全角空格

	return normalized, nil
}

// preprocessLines 预处理行：合并分散对齐的中文和孤立的数字行
func preprocessLines(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return ""
	}

	var result []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			result = append(result, line)
			continue
		}

		merged := trimmed
		nextIdx := i + 1

		for nextIdx < len(lines) {
			nextRaw := lines[nextIdx]
			nextTrimmed := strings.TrimSpace(nextRaw)

			if nextTrimmed == "" {
				nextIdx++
				continue
			}

			mergedUpdated := false

			// 规则 1: 短中文行合并 (无空格)
			if isShortChineseLine(merged) {
				if isShortChineseLine(nextTrimmed) || (len([]rune(nextTrimmed)) <= 4 && startsWithChinese(nextTrimmed)) {
					merged += nextTrimmed
					mergedUpdated = true
				}
			}

			// 规则 2: 纯数字行合并 (加空格)
			// 条件: 下一行是纯数字
			if !mergedUpdated {
				if numericLineRe.MatchString(nextTrimmed) {
					merged += " " + nextTrimmed
					mergedUpdated = true
				}
			}

			if mergedUpdated {
				i = nextIdx
				nextIdx++
			} else {
				break
			}
		}

		result = append(result, merged)
	}

	return strings.Join(result, "\n")
}

func isShortChineseLine(s string) bool {
	r := []rune(s)
	if len(r) == 0 || len(r) > 2 {
		return false
	}
	hasChinese := false
	for _, c := range r {
		if unicode.Is(unicode.Han, c) {
			hasChinese = true
			break
		}
	}
	return hasChinese
}

func startsWithChinese(s string) bool {
	r := []rune(s)
	if len(r) == 0 {
		return false
	}
	return unicode.Is(unicode.Han, r[0])
}

// normalizeText 合并PDF中的软换行
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
			if currentParagraph.Len() > 0 {
				result = append(result, currentParagraph.String())
				currentParagraph.Reset()
			}
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
				// 前一行不以标点结尾，合并
				needsSpace := needsSpaceBetween(prevContent, trimmed)

				if needsSpace {
					currentParagraph.WriteString(" ")
				}
				currentParagraph.WriteString(trimmed)
			}
		} else {
			currentParagraph.WriteString(trimmed)
		}

		if i == len(lines)-1 {
			if currentParagraph.Len() > 0 {
				result = append(result, currentParagraph.String())
			}
		}
	}

	return strings.Join(result, "\n")
}

// isLikelyTitle 判断是否可能是标题
func isLikelyTitle(line string) bool {
	runeCount := len([]rune(line))
	if runeCount > 50 {
		return false
	}
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
