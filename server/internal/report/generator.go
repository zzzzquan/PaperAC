package report

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"aigc-detector/server/internal/algo"
)

// ReportData 渲染模板所需的数据
type ReportData struct {
	Filename    string
	GeneratedAt string
	TotalRatio  string // e.g. "30.5%"
	BodyCount   int    // 正文句子数
	AICount     int    // AI句子数
	Sentences   []algo.SentenceResult
}

// GenerateHTML 生成 HTML 报告
func GenerateHTML(filename string, sentences []algo.SentenceResult, x float64) ([]byte, error) {
	// 计算实际的 AIGC 占比（只统计正文类型）
	var aiCount, bodyCount int
	for _, s := range sentences {
		if s.Type == "body" {
			bodyCount++
			if s.Label == "ai" {
				aiCount++
			}
		}
	}

	var ratioStr string
	if bodyCount > 0 {
		ratio := float64(aiCount) / float64(bodyCount)
		ratioStr = fmt.Sprintf("%.1f%%", ratio*100)
	} else {
		ratioStr = "0.0%"
	}

	data := ReportData{
		Filename:    filename,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		TotalRatio:  ratioStr,
		BodyCount:   bodyCount,
		AICount:     aiCount,
		Sentences:   sentences,
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 简单的内嵌 HTML 模板 - 纯文本堆叠格式
const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文鉴 - AIGC检测报告</title>
    <style>
        body {
            font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
            line-height: 1.8;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
            background: #f9f9f9;
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 1px solid #eee;
        }
        .header h1 {
            color: #1f2937;
            margin-bottom: 10px;
        }
        .meta {
            color: #6b7280;
            font-size: 0.9em;
        }
        .score-card {
            background: #fff;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.05);
            margin-bottom: 30px;
            text-align: center;
        }
        .score-value {
            font-size: 3em;
            font-weight: bold;
            color: #ef4444;
        }
        .score-label {
            color: #6b7280;
        }
        .score-detail {
            color: #9ca3af;
            font-size: 0.85em;
            margin-top: 8px;
        }
        .content {
            background: #fff;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.05);
            font-size: 16px;
            text-align: justify;
        }
        .ai {
            background-color: rgba(254, 202, 202, 0.5);
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>文鉴 AIGC 检测报告</h1>
        <div class="meta">
            <p>文件名: {{.Filename}}</p>
            <p>检测时间: {{.GeneratedAt}}</p>
        </div>
    </div>

    <div class="score-card">
        <div class="score-value">{{.TotalRatio}}</div>
        <div class="score-label">疑似 AIGC 生成比例</div>
        <div class="score-detail">检测句子数: {{.BodyCount}} | 疑似AI生成: {{.AICount}}</div>
    </div>

    <div class="content">{{range .Sentences}}<span class="{{.Label}}">{{.Text}}</span>{{end}}</div>
</body>
</html>
`
