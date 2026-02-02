package algo

import (
	"math/rand"
	"time"

	"aigc-detector/server/internal/splitter"
)

type SentenceResult struct {
	Text  string  `json:"text"`  // 句子文本
	Score float64 `json:"score"` // AIGC 概率 (0.0 - 1.0)
	Label string  `json:"label"` // "ai" (红) 或 "human" (绿) 或 "structural" (灰)
	Type  string  `json:"type"`  // 分段类型: "body", "title", "table", "list"
}

// Processor 处理文本并生成AIGC结果
type Processor struct {
	// 假如未来引入大模型或其他算法，可以在这里加字段
}

func NewProcessor() *Processor {
	return &Processor{}
}

// Process 处理句子列表，根据目标占比 x 生成结果（保持向后兼容）
// TODO: 用户反馈算法有优化空间，当前为 MVP 版本实现
func (p *Processor) Process(sentences []string, targetRatio float64) []SentenceResult {
	if len(sentences) == 0 {
		return nil
	}

	// 转换为 Segment 格式
	segments := make([]splitter.Segment, len(sentences))
	for i, s := range sentences {
		segments[i] = splitter.Segment{Text: s, Type: splitter.SegmentBody}
	}

	return p.ProcessWithSegments(segments, targetRatio)
}

// ProcessWithSegments 处理带类型的分段列表
// 只对正文类型进行AIGC检测，标题/表格/列表等结构化元素标记为 structural
func (p *Processor) ProcessWithSegments(segments []splitter.Segment, targetRatio float64) []SentenceResult {
	if len(segments) == 0 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())

	// 统计正文句子
	var bodyIndices []int
	for i, seg := range segments {
		if seg.Type == splitter.SegmentBody {
			bodyIndices = append(bodyIndices, i)
		}
	}

	bodyCount := len(bodyIndices)
	targetAICount := int(float64(bodyCount) * targetRatio)

	// 随机选择要标记为AI的正文句子
	rand.Shuffle(bodyCount, func(i, j int) {
		bodyIndices[i], bodyIndices[j] = bodyIndices[j], bodyIndices[i]
	})

	// 使用 map 记录哪些索引是 AI
	aiIndices := make(map[int]bool)
	for i := 0; i < targetAICount; i++ {
		aiIndices[bodyIndices[i]] = true
	}

	var results []SentenceResult

	for i, seg := range segments {
		var score float64
		var label string
		segType := string(seg.Type)

		// 非正文类型标记为 structural
		if seg.Type != splitter.SegmentBody {
			score = 0
			label = "structural"
		} else {
			isAI := aiIndices[i]
			if isAI {
				// 高分区间 (0.8 ~ 0.99)
				score = 0.8 + rand.Float64()*0.19
				label = "ai"
			} else {
				// 低分区间 (0.0 ~ 0.2)
				score = rand.Float64() * 0.2
				label = "human"
			}
		}

		results = append(results, SentenceResult{
			Text:  seg.Text,
			Score: score,
			Label: label,
			Type:  segType,
		})
	}

	return results
}
