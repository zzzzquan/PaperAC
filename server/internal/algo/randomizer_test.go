package algo

import (
	"fmt"
	"math"
	"testing"
	"unicode/utf8"

	"aigc-detector/server/internal/splitter"
)

func TestProcessWithSegments_TargetRatioRange(t *testing.T) {
	processor := NewProcessor()
	segments := make([]splitter.Segment, 100)
	for i := 0; i < 100; i++ {
		segments[i] = splitter.Segment{
			Text: "这是一个足够长的测试句子，超过了十个字符。",
			Type: splitter.SegmentBody,
		}
	}

	// 运行多次，检查目标比例范围
	for i := 0; i < 100; i++ {
		result := processor.ProcessWithSegments(segments)
		if result.TargetRatio < 7.0 || result.TargetRatio > 15.0 {
			t.Errorf("目标比例 %.1f 超出范围 [7.0, 15.0]", result.TargetRatio)
		}
	}
}

func TestProcessWithSegments_FilterShortSentences(t *testing.T) {
	processor := NewProcessor()
	segments := []splitter.Segment{
		{Text: "短句", Type: splitter.SegmentBody},             // Should be filtered
		{Text: "这也是短句", Type: splitter.SegmentBody},          // Should be filtered
		{Text: "这是一个长句子，应该被考虑。", Type: splitter.SegmentBody}, // Should use
	}

	result := processor.ProcessWithSegments(segments)

	for i, s := range result.Sentences {
		charCount := utf8.RuneCountInString(s.Text)
		if charCount < MinSentenceChars {
			if s.Label == "ai" {
				t.Errorf("第 %d 个句子太短 (%d chars) 却被标记为AI", i, charCount)
			}
		}
	}
}

func TestProcessWithSegments_ClusterSelection(t *testing.T) {
	processor := NewProcessor()
	// 创建大量句子，足以显示分布
	count := 200
	segments := make([]splitter.Segment, count)
	for i := 0; i < count; i++ {
		segments[i] = splitter.Segment{
			Text: fmt.Sprintf("这是第 %d 个测试句子，长度足够长。", i),
			Type: splitter.SegmentBody,
		}
	}

	result := processor.ProcessWithSegments(segments)

	// 收集被标记为AI的索引
	var aiIndices []int
	for i, s := range result.Sentences {
		if s.Label == "ai" {
			aiIndices = append(aiIndices, i)
		}
	}

	if len(aiIndices) == 0 {
		return // 比例太低可能什么都没选到
	}

	// 验证集中性：计算平均相邻距离
	// 完美的集中应该是 indices: 10, 11, 12, 13... 平均距离=1
	// 分散的可能是: 10, 50, 90... 平均距离很大

	if len(aiIndices) > 1 {
		totalDist := 0
		for i := 0; i < len(aiIndices)-1; i++ {
			dist := aiIndices[i+1] - aiIndices[i]
			totalDist += dist
		}
		avgDist := float64(totalDist) / float64(len(aiIndices)-1)

		// 允许一定的空隙（可能因为随机左右扩展和跳过短句），但应该比较小
		// 实际上由于扩展算法是连续的，除非中间有被过滤的短句，否则距离应该是1
		// 如果是随机分散，平均距离大约是 Total / SelectionCount

		expectedDistributedDist := float64(count) / float64(len(aiIndices))

		t.Logf("选中的句子数量: %d", len(aiIndices))
		t.Logf("平均相邻距离: %.2f", avgDist)
		t.Logf("如果是随机分散，预计平均距离: %.2f", expectedDistributedDist)

		if avgDist > 2.0 && avgDist > expectedDistributedDist/2 {
			// 如果平均距离很大，说明很分散（考虑到中间可能有短句，阈值设宽一点）
			t.Errorf("选句不够集中，平均距离 %.2f 偏大", avgDist)
		}
	}
}

func TestProcessWithSegments_IgnoreNonBody(t *testing.T) {
	processor := NewProcessor()
	segments := []splitter.Segment{
		{Text: "标题", Type: splitter.SegmentTitle},
		{Text: "这是一个很长的正文句子。", Type: splitter.SegmentBody},
		{Text: "1. 列表项", Type: splitter.SegmentList},
		{Text: "封面信息", Type: splitter.SegmentMeta},
	}

	result := processor.ProcessWithSegments(segments)

	for _, s := range result.Sentences {
		if s.Type != "body" {
			if s.Label != "structural" {
				t.Errorf("非正文类型 %s 应该标为 structural，实际为 %s", s.Type, s.Label)
			}
			if s.Score != 0 {
				t.Errorf("非正文类型 %s 分数应为0，实际为 %.2f", s.Type, s.Score)
			}
		}
	}
}

func TestSelectClusteredSentences_Logic(t *testing.T) {
	// 直接测试内部逻辑可能需要导出函数或者把测试放在同一个包下
	// 这里通过 ProcessWithSegments 间接测试即可
}

// 辅助函数：计算字数比例
func calculateActualRatio(sentences []SentenceResult) float64 {
	var aiChars, totalChars int
	for _, s := range sentences {
		if s.Type == "body" {
			c := utf8.RuneCountInString(s.Text)
			totalChars += c
			if s.Label == "ai" {
				aiChars += c
			}
		}
	}
	if totalChars == 0 {
		return 0
	}
	return float64(aiChars) / float64(totalChars) * 100.0
}

func TestProcessor_RatioAccuracy(t *testing.T) {
	processor := NewProcessor()
	// 构建足够多的数据以减少离散误差
	count := 500
	segments := make([]splitter.Segment, count)
	for i := 0; i < count; i++ {
		segments[i] = splitter.Segment{
			Text: "这是一个标准的十个字句子。", // 12 chars
			Type: splitter.SegmentBody,
		}
	}

	result := processor.ProcessWithSegments(segments)

	// 允许 5% 的绝对误差（因为必须选完整句子，可能会超一点或少一点）
	diff := math.Abs(result.ActualRatio - result.TargetRatio)

	t.Logf("目标比例: %.2f%%, 实际比例: %.2f%%", result.TargetRatio, result.ActualRatio)

	if diff > 5.0 {
		t.Errorf("实际比例偏差过大: 目标 %.2f%%, 实际 %.2f%%", result.TargetRatio, result.ActualRatio)
	}
}
