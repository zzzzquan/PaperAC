package algo

import (
	"math/rand"
	"sort"
	"time"
	"unicode"
	"unicode/utf8"

	"aigc-detector/server/internal/splitter"
)

// 最小句子字符数（少于此值的句子不参与AIGC选取）
// 修改：从 10 提升到 20，且仅统计汉字字符（不含标点）
const MinSentenceChars = 20

type SentenceResult struct {
	Text  string  `json:"text"`  // 句子文本
	Score float64 `json:"score"` // AIGC 概率 (0.0 - 1.0)
	Label string  `json:"label"` // "ai" (红) 或 "human" (绿) 或 "structural" (灰)
	Type  string  `json:"type"`  // 分段类型: "body", "title", "table", "list", "meta"
}

// ProcessResult 处理结果，包含生成的AIGC比例
type ProcessResult struct {
	Sentences   []SentenceResult
	TargetRatio float64 // 随机生成的目标AIGC比例 (7.0 ~ 15.0)
	ActualRatio float64 // 实际达到的AIGC比例
}

// Processor 处理文本并生成AIGC结果
type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

// Process 处理句子列表（保持向后兼容）
func (p *Processor) Process(sentences []string, _ float64) []SentenceResult {
	if len(sentences) == 0 {
		return nil
	}

	segments := make([]splitter.Segment, len(sentences))
	for i, s := range sentences {
		segments[i] = splitter.Segment{Text: s, Type: splitter.SegmentBody}
	}

	result := p.ProcessWithSegments(segments)
	return result.Sentences
}

// ProcessWithSegments 处理带类型的分段列表
// 新规则：
// 1. AIGC比例随机生成 7.0~15.0%
// 2. 按字数计算比例
// 3. 排除少于20有效字符的句子（仅统计汉字，排除标点）
// 4. "分布式块+散点"策略
func (p *Processor) ProcessWithSegments(segments []splitter.Segment) ProcessResult {
	if len(segments) == 0 {
		return ProcessResult{}
	}

	rand.Seed(time.Now().UnixNano())

	// 随机生成目标AIGC比例：7.0% ~ 15.0%，保留一位小数
	targetRatio := float64(70+rand.Intn(81)) / 10.0 // 7.0 ~ 15.0

	// 收集可选句子（正文类型 + 有效字符数>=20）
	var candidates []candidate
	var totalBodyChars int

	for i, seg := range segments {
		if seg.Type == splitter.SegmentBody {
			// 计算所有字符数用于显示/计算比例基数？
			// 通常AIGC比例是基于正文总字数的。是否剔除标点？
			// 这里我们为了保持简单，totalBodyChars 仍然统计总字符数（utf8 count），
			// 但是筛选 candidates 时使用 validCharCount (汉字数)。
			// 或者为了准确，应该都统一标准。
			// 用户只说 "选取的句子汉字数（不包含标点符号）改成需要大于等于20字"
			// 也就是 filter condition changed.

			charCount := utf8.RuneCountInString(seg.Text)
			totalBodyChars += charCount

			validCount := countChineseChars(seg.Text)
			if validCount >= MinSentenceChars {
				candidates = append(candidates, candidate{index: i, charCount: charCount}) // 仍然存储总字数用于配额计算
			}
		}
	}

	// 计算目标字数
	targetChars := int(float64(totalBodyChars) * targetRatio / 100.0)

	// 使用分布式选择策略
	aiIndices := selectDistributedSentences(candidates, targetChars)

	// 计算实际字数
	var actualAIChars int
	aiSet := make(map[int]bool)
	for _, idx := range aiIndices {
		aiSet[idx] = true
		actualAIChars += utf8.RuneCountInString(segments[idx].Text)
	}

	actualRatio := 0.0
	if totalBodyChars > 0 {
		actualRatio = float64(actualAIChars) / float64(totalBodyChars) * 100.0
	}

	// 生成结果
	var results []SentenceResult
	for i, seg := range segments {
		var score float64
		var label string
		segType := string(seg.Type)

		if seg.Type != splitter.SegmentBody {
			score = 0
			label = "structural"
		} else {
			// 对于是否参与检测显示，如果不满足20字，自然不可能被 label="ai" (因为不在 candidates 里)
			// 但是否标记为 "human" 还是 "structural" (gray)?
			// 保持 label="human" (green)，因为它是正文，只是太短没机会变红。

			// validCount := countChineseChars(seg.Text)
			// check if in aiSet
			if aiSet[i] {
				// 被选中为AI
				score = 0.8 + rand.Float64()*0.19
				label = "ai"
			} else {
				// 未选中
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

	return ProcessResult{
		Sentences:   results,
		TargetRatio: targetRatio,
		ActualRatio: actualRatio,
	}
}

// countChineseChars 统计字符串中的汉字数量（排除标点等）
func countChineseChars(s string) int {
	count := 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			count++
		}
	}
	return count
}

// selectDistributedSentences 分布式选择策略
func selectDistributedSentences(candidates []candidate, targetChars int) []int {
	if len(candidates) == 0 || targetChars <= 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].index < candidates[j].index
	})

	used := make(map[int]bool)
	var selected []int
	currentChars := 0

	// 参数配置
	const blockRatio = 0.8
	blockQuota := int(float64(targetChars) * blockRatio)
	minBlockSize := 3
	maxBlockSize := 6

	maxRetries := len(candidates) * 2
	retries := 0

	for currentChars < blockQuota && retries < maxRetries {
		retries++
		startIdx := rand.Intn(len(candidates))
		c := candidates[startIdx]

		if used[c.index] {
			continue
		}

		blockSize := minBlockSize + rand.Intn(maxBlockSize-minBlockSize+1)
		var block []int
		blockChars := 0

		currSliceIdx := startIdx
		validBlock := true

		for k := 0; k < blockSize; k++ {
			if currSliceIdx >= len(candidates) {
				break
			}

			currCandidate := candidates[currSliceIdx]

			if used[currCandidate.index] {
				break
			}
			if k > 0 {
				prevCandidate := candidates[currSliceIdx-1]
				if currCandidate.index != prevCandidate.index+1 {
					break
				}
			}

			if currentChars+blockChars+currCandidate.charCount > targetChars {
				validBlock = false
				break
			}

			block = append(block, currCandidate.index)
			blockChars += currCandidate.charCount
			currSliceIdx++
		}

		if len(block) >= minBlockSize || (len(block) > 0 && blockQuota-currentChars < 200) {
			if validBlock {
				for _, idx := range block {
					used[idx] = true
					selected = append(selected, idx)
				}
				currentChars += blockChars
			}
		}
	}

	// Phase 2: Scatter Selection
	perm := rand.Perm(len(candidates))
	for _, i := range perm {
		if currentChars >= targetChars {
			break
		}
		c := candidates[i]
		if !used[c.index] {
			if currentChars+c.charCount <= targetChars+100 {
				used[c.index] = true
				selected = append(selected, c.index)
				currentChars += c.charCount
			}
		}
	}

	return selected
}

type candidate struct {
	index     int
	charCount int
}
