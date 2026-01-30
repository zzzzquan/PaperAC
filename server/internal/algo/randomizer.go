package algo

import (
	"math/rand"
	"time"
)

type SentenceResult struct {
	Text  string  `json:"text"`  // 句子文本
	Score float64 `json:"score"` // AIGC 概率 (0.0 - 1.0)
	Label string  `json:"label"` // "ai" (红) 或 "human" (绿)
}

// Processor 处理文本并生成AIGC结果
type Processor struct {
	// 假如未来引入大模型或其他算法，可以在这里加字段
}

func NewProcessor() *Processor {
	return &Processor{}
}

// Process 处理句子列表，根据目标占比 x 生成结果
// TODO: 用户反馈算法有优化空间，当前为 MVP 版本实现
func (p *Processor) Process(sentences []string, targetRatio float64) []SentenceResult {
	if len(sentences) == 0 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())

	total := len(sentences)
	targetAICount := int(float64(total) * targetRatio)

	// 创建一个索引列表 [0, 1, 2, ..., total-1]
	indices := make([]int, total)
	for i := 0; i < total; i++ {
		indices[i] = i
	}

	// 随机打乱索引，前 targetAICount 个被选中为 AI 句
	rand.Shuffle(total, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// 使用 map 记录哪些索引是 AI
	aiIndices := make(map[int]bool)
	for i := 0; i < targetAICount; i++ {
		aiIndices[indices[i]] = true
	}

	var results []SentenceResult

	for i, text := range sentences {
		isAI := aiIndices[i]
		var score float64
		var label string

		if isAI {
			// 高分区间 (0.8 ~ 0.99)
			score = 0.8 + rand.Float64()*0.19
			label = "ai"
		} else {
			// 低分区间 (0.0 ~ 0.2)
			// 为了增加真实感，偶尔混入一点中等分数? 暂时先简单点
			score = rand.Float64() * 0.2
			label = "human"
		}

		results = append(results, SentenceResult{
			Text:  text,
			Score: score,
			Label: label,
		})
	}

	return results
}
