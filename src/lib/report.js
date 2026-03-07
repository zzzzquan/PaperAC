export function previewReport(filename, sentences) {
    const html = generateHTML(filename, sentences);
    const blob = new Blob([html], { type: 'text/html; charset=utf-8' });
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
}

function generateHTML(filename, sentences) {
    let aiChars = 0;
    let totalBodyChars = 0;

    for (const s of sentences) {
        if (s.type === 'body') {
            const charCount = [...s.text].length;
            totalBodyChars += charCount;
            if (s.label === 'ai') {
                aiChars += charCount;
            }
        }
    }

    const ratioStr = totalBodyChars > 0 ? (aiChars / totalBodyChars * 100).toFixed(1) + '%' : '0.0%';

    const generatedAt = new Date().toLocaleString('zh-CN', {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit', second: '2-digit',
        hour12: false
    });

    const contentHTML = sentences.map(s => {
        if (s.type === 'newline') {
            return '<br>';
        }
        const escaped = escapeHTML(s.text);
        return `<span class="${s.label}">${escaped}</span>`;
    }).join('');

    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文鉴 - AIGC检测报告</title>
    <style>
        body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif; line-height: 1.8; color: #333; max-width: 800px; margin: 0 auto; padding: 40px 20px; background: #f9f9f9; }
        .header { text-align: center; margin-bottom: 40px; padding-bottom: 20px; border-bottom: 1px solid #eee; }
        .header h1 { color: #1f2937; margin-bottom: 10px; }
        .meta { color: #6b7280; font-size: 0.9em; }
        .score-card { background: #fff; padding: 20px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); margin-bottom: 30px; text-align: center; }
        .score-value { font-size: 3em; font-weight: bold; color: #ef4444; }
        .score-label { color: #6b7280; }
        .score-detail { color: #9ca3af; font-size: 0.85em; margin-top: 8px; }
        .content { background: #fff; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); font-size: 16px; text-align: justify; }
        .ai { background-color: rgba(254, 202, 202, 0.5); }
    </style>
</head>
<body>
    <div class="header">
        <h1>文鉴 AIGC 检测报告</h1>
        <div class="meta">
            <p>文件名: ${escapeHTML(filename)}</p>
            <p>检测时间: ${generatedAt}</p>
        </div>
    </div>
    <div class="score-card">
        <div class="score-value">${ratioStr}</div>
        <div class="score-label">疑似 AIGC 生成比例</div>
        <div class="score-detail">正文总字数: ${totalBodyChars} | 疑似AI生成字数: ${aiChars}</div>
    </div>
    <div class="content">${contentHTML}</div>
</body>
</html>`;
}

function escapeHTML(str) {
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;');
}
