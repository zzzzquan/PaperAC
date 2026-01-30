package splitter

import (
	"regexp"
	"strings"
)

// Split 将长文本切分为句子
func Split(text string) []string {
	// 简单的正则切分，匹配中英文句号、问号、叹号、分号
	// 注意：这里只是简单的MVP实现，后续可以优化（比如保留标点符号在句尾）
	re := regexp.MustCompile(`([。？！；!?;]+)`)
	
	// 使用 Wrap 逻辑保留标点符号
	// 这里简化处理：直接 split，标点可能会丢失或附着在下一句
	// 更好的做法是 ReplaceAllStringFunc 或者手动遍历
	
	// 临时方案：把标点替换成 "标点\n"，然后按行切分
	temp := re.ReplaceAllString(text, "$1\n")
	
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
