package splitter

import (
	"testing"
)

func TestSplitWithStructure(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Segment
	}{
		{
			name:  "识别章节标题",
			input: "第一章 绪论\n本章介绍研究背景。",
			expected: []Segment{
				{Text: "第一章 绪论", Type: SegmentTitle},
				{Text: "本章介绍研究背景。", Type: SegmentBody},
			},
		},
		{
			name:  "识别列表项",
			input: "主要内容包括：\n1. 第一点\n2. 第二点",
			expected: []Segment{
				{Text: "主要内容包括：", Type: SegmentBody},
				{Text: "1. 第一点", Type: SegmentList},
				{Text: "2. 第二点", Type: SegmentList},
			},
		},
		{
			name:  "正文分句",
			input: "这是第一句。这是第二句！这是第三句？",
			expected: []Segment{
				{Text: "这是第一句。", Type: SegmentBody},
				{Text: "这是第二句！", Type: SegmentBody},
				{Text: "这是第三句？", Type: SegmentBody},
			},
		},
		{
			name:  "识别摘要标题",
			input: "摘 要\n本文研究了人工智能。",
			expected: []Segment{
				{Text: "摘 要", Type: SegmentTitle},
				{Text: "本文研究了人工智能。", Type: SegmentBody},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitWithStructure(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("长度不匹配: 期望 %d, 实际 %d", len(tt.expected), len(result))
				t.Errorf("实际结果: %+v", result)
				return
			}
			for i, seg := range result {
				if seg.Text != tt.expected[i].Text || seg.Type != tt.expected[i].Type {
					t.Errorf("第%d个分段不匹配:\n期望: %+v\n实际: %+v", i, tt.expected[i], seg)
				}
			}
		})
	}
}

func TestDetectSegmentType(t *testing.T) {
	tests := []struct {
		input    string
		expected SegmentType
	}{
		{"第一章 绪论", SegmentTitle},
		{"摘要", SegmentTitle},
		{"Abstract", SegmentTitle},
		{"参考文献", SegmentTitle},
		{"1. 第一项内容", SegmentList},
		{"• 项目符号", SegmentList},
		{"- 短横线列表", SegmentList},
		{"这是一个普通的正文句子。", SegmentBody},
		{"研究方法与技术路线", SegmentTitle},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := detectSegmentType(tt.input)
			if result != tt.expected {
				t.Errorf("detectSegmentType(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplit_BackwardCompatibility(t *testing.T) {
	// 测试原有 Split 函数仍能正常工作
	input := "这是第一句。这是第二句！"
	result := Split(input)

	if len(result) != 2 {
		t.Errorf("Split 结果长度错误: 期望 2, 实际 %d", len(result))
	}
	if result[0] != "这是第一句。" {
		t.Errorf("第一句错误: %s", result[0])
	}
	if result[1] != "这是第二句！" {
		t.Errorf("第二句错误: %s", result[1])
	}
}

func TestHasSpecialCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"这是正常的中文句子。", false},
		{"包含《书名号》的句子。", false},
		{"包含、顿号的句子。", false},
		{"公式a=b+c不应被检测。", true},
		{"含有#井号的句子。", true},
		{"含有$美元符的句子。", true},
		{"含有希腊字母α的句子。", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := hasSpecialCharacters(tt.input)
			if result != tt.expected {
				t.Errorf("hasSpecialCharacters(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}
