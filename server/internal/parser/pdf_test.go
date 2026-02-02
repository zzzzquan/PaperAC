package parser

import (
	"testing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "合并中文软换行",
			input:    "这是一个很长的句子，因为PDF页面宽度有限所以\n自动换行到了下一行继续显示。",
			expected: "这是一个很长的句子，因为PDF页面宽度有限所以自动换行到了下一行继续显示。",
		},
		{
			name:     "保留以标点结尾的段落",
			input:    "第一段结束。\n第二段开始。",
			expected: "第一段结束。\n第二段开始。",
		},
		{
			name:     "识别章节标题",
			input:    "第一章 绪论\n本章主要介绍研究背景。",
			expected: "第一章 绪论\n本章主要介绍研究背景。",
		},
		{
			name:     "识别摘要标题",
			input:    "摘 要\n本文研究了人工智能技术。",
			expected: "摘 要\n本文研究了人工智能技术。",
		},
		{
			name:     "英文需要空格",
			input:    "This is a long\nsentence.",
			expected: "This is a long sentence.",
		},
		{
			name:     "中文不需要空格",
			input:    "这是一个长\n句子。",
			expected: "这是一个长句子。",
		},
		{
			name:     "空行分隔段落",
			input:    "第一段内容。\n\n第二段内容。",
			expected: "第一段内容。\n第二段内容。",
		},
		{
			name:     "多行合并",
			input:    "这是第一行\n这是第二行\n这是第三行。",
			expected: "这是第一行这是第二行这是第三行。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeText(tt.input)
			if result != tt.expected {
				t.Errorf("\n输入: %q\n期望: %q\n实际: %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestIsLikelyTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"第一章 绪论", true},
		{"第二节 研究方法", true},
		{"摘要", true},
		{"摘 要", true},
		{"Abstract", true},
		{"引言", true},
		{"结论", true},
		{"参考文献", true},
		{"致谢", true},
		{"附录", true},
		{"这是一个普通的句子。", false},
		{"这是一个很长很长的句子，超过了五十个字符的限制所以不会被识别为标题。", false},
		{"这是一个不以标点结尾的", false}, // 以连接词结尾
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isLikelyTitle(tt.input)
			if result != tt.expected {
				t.Errorf("isLikelyTitle(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNeedsSpaceBetween(t *testing.T) {
	tests := []struct {
		prev     string
		next     string
		expected bool
	}{
		{"Hello", "world", true},
		{"你好", "世界", false},
		{"Hello", "世界", false},
		{"你好", "world", false},
		{"test123", "abc", true},
		{"", "abc", false},
		{"abc", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.prev+"_"+tt.next, func(t *testing.T) {
			result := needsSpaceBetween(tt.prev, tt.next)
			if result != tt.expected {
				t.Errorf("needsSpaceBetween(%q, %q) = %v, 期望 %v", tt.prev, tt.next, result, tt.expected)
			}
		})
	}
}
