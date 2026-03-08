export function downloadReport(filename, sentences) {
    const html = generateHTML(filename, sentences);
    // Use hidden iframe to trigger print without opening a new tab
    const iframe = document.createElement('iframe');
    iframe.style.position = 'fixed';
    iframe.style.right = '0';
    iframe.style.bottom = '0';
    iframe.style.width = '0';
    iframe.style.height = '0';
    iframe.style.border = 'none';
    iframe.style.opacity = '0';
    document.body.appendChild(iframe);

    const iframeDoc = iframe.contentDocument || iframe.contentWindow.document;
    iframeDoc.open();
    iframeDoc.write(html);
    iframeDoc.close();

    iframe.onload = () => {
        setTimeout(() => {
            iframe.contentWindow.focus();
            iframe.contentWindow.print();
            // Clean up iframe after printing
            setTimeout(() => {
                document.body.removeChild(iframe);
            }, 1000);
        }, 400);
    };
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

    const aiRatioNum = totalBodyChars > 0 ? (aiChars / totalBodyChars * 100) : 0;
    const ratioStr = aiRatioNum.toFixed(1) + '%';
    const r = 60;
    const circ = 2 * Math.PI * r;
    const dash = (aiRatioNum / 100) * circ;

    // 计算全部文字字符数（包括标题、列表等所有可见文字，排除换行）
    let totalAllChars = 0;
    for (const s of sentences) {
        if (s.type !== 'newline' && s.text) {
            totalAllChars += [...s.text].length;
        }
    }

    let segmentsHTML = '';
    for (const s of sentences) {
        if (s.type !== 'newline' && s.text) {
            const charCount = [...s.text].length;
            const widthPct = totalAllChars > 0 ? (charCount / totalAllChars) * 100 : 0;
            segmentsHTML += `<div style="flex: 0 0 ${widthPct}%; background-color: ${s.label === 'ai' ? '#ef4444' : 'transparent'};"></div>`;
        }
    }

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
    <title>PaperAC - AIGC检测报告单</title>
    <style>
        body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif; line-height: 1.8; color: #333; max-width: 800px; margin: 0 auto; padding: 40px 20px; background: #f9f9f9; position: relative; }
        .watermark { position: fixed; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 9999; background: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' version='1.1' height='250px' width='400px'><text x='200' y='125' fill='%23000000' font-size='30' font-family='Arial' transform='rotate(-30 200 125)' text-anchor='middle' opacity='0.09'>https://www.paperac.com/</text></svg>") repeat; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; padding-bottom: 20px; border-bottom: 2px solid #ef4444; }
        .header-logo { display: flex; align-items: center; gap: 8px; font-weight: bold; font-size: 1.2em; color: #1f2937; max-width: 380px; }
        .header-title { font-size: 1.8em; font-weight: bold; color: #1f2937; text-align: center; margin-top: 20px; margin-bottom: 10px; }
        .meta { color: #6b7280; font-size: 0.9em; text-align: center; margin-bottom: 30px; }
        .dashboard { display: flex; flex-direction: column; gap: 20px; margin-bottom: 30px; }
        .ring-chart-card { display: flex; align-items: center; justify-content: center; gap: 50px; background: #fff; padding: 30px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
        .ring-chart-container { position: relative; width: 160px; height: 160px; }
        .ring-chart-container svg { width: 100%; height: 100%; transform: rotate(-90deg); }
        .ring-center { position: absolute; top: 0; left: 0; width: 100%; height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; }
        .ring-value { font-size: 1.6em; font-weight: bold; line-height: 1; margin-bottom: 6px; }
        .ring-label { font-size: 1em; color: #6b7280; }
        .stats-info { display: flex; flex-direction: column; gap: 16px; font-size: 1.2em; color: #4b5563; }
        .stats-info p { margin: 0; }
        .dist-chart-card { background: #fff; padding: 30px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
        .dist-bar { display: flex; width: 100%; height: 24px; background: #e5e7eb; border-radius: 2px; margin-bottom: 8px; overflow: hidden; /* print compatibility */ -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        .dist-labels { display: flex; justify-content: space-between; font-size: 0.9em; color: #9ca3af; margin-bottom: 20px; }
        .dist-legend { display: flex; justify-content: center; gap: 30px; font-size: 0.9em; color: #374151; }
        .legend-item { display: flex; align-items: center; gap: 6px; }
        .legend-color { width: 12px; height: 12px; border-radius: 2px; /* print compatibility */ -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        .content { background: #fff; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); font-size: 16px; text-align: justify; }
        .ai { background-color: rgba(254, 202, 202, 0.5); border-bottom: 2px solid #ef4444; padding-bottom: 2px; /* print compatibility */ -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        @media print {
            body { background: #fff; margin: 0; padding: 0; max-width: 100%; }
            .watermark { position: fixed; top: 0; left: 0; width: 100%; height: 100%; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
            .score-card, .content, .ring-chart-card, .dist-chart-card { box-shadow: none; padding: 0; margin-bottom: 20px; }
            .dist-bar div { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        }
    </style>
</head>
<body>
    <div class="watermark"></div>
    <div class="header">
        <div class="header-logo">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 300" width="360" height="108">
                <defs><style>.text-title{font-family:"Nunito","Quicksand","Baloo 2","Chalkboard SE","Comic Sans MS",sans-serif;font-size:140px;font-weight:800;fill:#374151;letter-spacing:-1px;}.accent{fill:#f28219;}.divider{stroke:#CBD5E1;stroke-width:6px;stroke-linecap:round;stroke-dasharray:10,15;}</style></defs>
                <g transform="translate(40, 40)">
                    <svg x="0" y="0" width="220" height="220" viewBox="180 170 680 680">
                        <g><path fill="#f28219" d="M 574.5,537.5 C 575.675,537.719 576.675,537.386 577.5,536.5C 575.364,535.207 573.364,535.207 571.5,536.5C 545.139,536.103 522.973,545.436 505,564.5C 489.838,585.804 487.671,608.471 498.5,632.5C 485.021,629.217 475.521,620.884 470,607.5C 459.809,572.39 463.975,539.057 482.5,507.5C 467.039,513.805 456.206,525.139 450,541.5C 447.928,547.048 446.262,552.715 445,558.5C 442.808,578.038 439.475,597.372 435,616.5C 429.274,633.06 418.107,644.06 401.5,649.5C 401.167,649.333 400.833,649.167 400.5,649C 410.865,629.101 414.698,607.935 412,585.5C 410.945,576.615 408.945,567.948 406,559.5C 397.937,539.311 390.271,518.978 383,498.5C 374.502,456.659 386.669,422.159 419.5,395C 445.996,375.273 475.662,362.939 508.5,358C 484.742,350.702 460.742,349.702 436.5,355C 392.138,369.507 366.971,400.007 361,446.5C 360.333,455.5 360.333,464.5 361,473.5C 362.592,483.927 364.259,494.26 366,504.5C 337.593,453.993 331.593,400.993 348,345.5C 353.606,330.281 361.273,316.281 371,303.5C 369.138,272.466 375.805,243.466 391,216.5C 392.113,214.671 393.613,213.837 395.5,214C 408.157,226.816 419.49,240.649 429.5,255.5C 435.137,256.663 440.804,256.83 446.5,256C 456.629,242.201 468.963,230.868 483.5,222C 486.971,219.928 490.638,218.428 494.5,217.5C 495.118,240.297 497.118,262.964 500.5,285.5C 503.817,288.806 506.651,292.472 509,296.5C 512.093,304.788 515.76,312.788 520,320.5C 529.272,330.537 540.106,338.204 552.5,343.5C 546.78,352.034 539.113,358.201 529.5,362C 516.56,367.068 503.227,370.568 489.5,372.5C 493.217,390.761 502.217,405.928 516.5,418C 523.479,423.823 530.812,429.156 538.5,434C 561.659,446.079 584.659,458.412 607.5,471C 629.683,484.339 647.849,501.839 662,523.5C 678.683,552.002 683.349,582.336 676,614.5C 668.072,636.393 652.572,649.226 629.5,653C 610.628,654.847 591.961,653.513 573.5,649C 556.471,644.268 539.471,639.435 522.5,634.5C 506.243,601.698 512.577,574.198 541.5,552C 551.559,544.969 562.559,540.136 574.5,537.5 Z"/></g>
                        <g><path fill="#f28219" d="M 571.5,536.5 C 573.364,535.207 575.364,535.207 577.5,536.5C 576.675,537.386 575.675,537.719 574.5,537.5C 573.791,536.596 572.791,536.263 571.5,536.5 Z"/></g>
                        <g><path fill="#f28219" d="M 694.5,591.5 C 730.085,623.096 739.252,661.096 722,705.5C 709.833,728.333 692.333,745.833 669.5,758C 642.256,772.456 613.256,780.79 582.5,783C 567.518,784.298 552.518,784.632 537.5,784C 507.773,782.061 478.106,779.394 448.5,776C 415.457,772.15 383.457,776.15 352.5,788C 341.231,792.878 331.064,799.378 322,807.5C 318.649,760.208 334.815,721.041 370.5,690C 411.764,655.033 459.098,642.7 512.5,653C 532.872,658.176 553.205,663.51 573.5,669C 600.847,675.879 627.847,674.879 654.5,666C 686.002,651.176 699.335,626.342 694.5,591.5 Z"/></g>
                        <g><path fill="#faf8f7" d="M 429.5,688.5 C 449.142,688.395 467.809,692.561 485.5,701C 469.908,702.97 455.908,708.636 443.5,718C 440.269,720.727 437.603,723.894 435.5,727.5C 446.422,728.918 457.089,731.418 467.5,735C 483.318,742.076 498.985,749.41 514.5,757C 523.232,760.799 532.232,763.799 541.5,766C 511.207,762.356 480.874,758.689 450.5,755C 415.805,752.059 382.805,758.059 351.5,773C 346.074,776.129 340.907,779.629 336,783.5C 342.622,746.902 361.789,719.069 393.5,700C 405.011,694.217 417.011,690.384 429.5,688.5 Z"/></g>
                        <g><path fill="#f28219" d="M 681.5,715.5 C 684.778,716.941 687.944,718.608 691,720.5C 691.756,722.809 691.256,724.809 689.5,726.5C 683.299,728.261 678.632,731.928 675.5,737.5C 672.843,738.259 670.177,738.926 667.5,739.5C 665.047,735.796 662.881,735.796 661,739.5C 659.676,738.12 658.343,737.786 657,738.5C 655.336,735.711 655.17,732.877 656.5,730C 657.449,729.383 658.282,729.549 659,730.5C 662.05,726.905 665.216,723.405 668.5,720C 672.833,718.181 677.167,716.681 681.5,715.5 Z"/></g>
                    </svg>
                    <line x1="300" y1="40" x2="300" y2="180" class="divider" />
                    <text x="360" y="165" class="text-title">Paper<tspan class="accent">AC</tspan></text>
                    <circle cx="880" cy="55" r="8" fill="#f28219" opacity="0.85" />
                    <circle cx="910" cy="40" r="4" fill="none" stroke="#f28219" stroke-width="2" opacity="0.6" />
                    <circle cx="855" cy="45" r="3" fill="#f28219" opacity="0.4" />
                    <path d="M 375 190 Q 630 205 885 185" fill="none" stroke="#f28219" stroke-width="6" stroke-linecap="round" opacity="0.8" />
                </g>
            </svg>
        </div>
        <div style="font-size: 0.85em; color: #9ca3af;">PaperAC</div>
    </div>
    
    <div class="header-title">AIGC检测报告单</div>

    <div class="meta">
        <p>文件名: ${escapeHTML(filename)}</p>
        <p>检测时间: ${generatedAt}</p>
    </div>
    
    <div class="dashboard">
        <div class="ring-chart-card">
            <div class="ring-chart-container">
                <svg width="160" height="160" viewBox="0 0 160 160">
                    <circle cx="80" cy="80" r="60" fill="none" stroke="#e5e7eb" stroke-width="20" />
                    <circle cx="80" cy="80" r="60" fill="none" stroke="#ef4444" stroke-width="20"
                            stroke-dasharray="${dash} ${circ}" stroke-dashoffset="0" />
                </svg>
                <div class="ring-center">
                    <div class="ring-value" style="color: ${aiRatioNum > 0 ? '#ef4444' : '#6b7280'};">${ratioStr}</div>
                    <div class="ring-label">AI特征值</div>
                </div>
            </div>
            <div class="stats-info">
                <p>AI特征值: <span style="font-size:1.3em; font-weight:bold; color:#ef4444; margin-left:8px;">${ratioStr}</span></p>
                <p>AI特征字符数: <span style="font-size:1.3em; font-weight:bold; color:#ef4444; margin-left:8px;">${aiChars}</span></p>
                <p>总字符数: <span style="font-size:1.3em; font-weight:bold; color:#333; margin-left:8px;">${totalBodyChars}</span></p>
            </div>
        </div>
        
        <div class="dist-chart-card">
            <div class="dist-bar">
                ${segmentsHTML}
            </div>
            <div class="dist-labels">
                <span>0</span>
                <span>${totalAllChars}</span>
            </div>
            <div class="dist-legend">
                <span class="legend-item"><span class="legend-color" style="background:#ef4444; -webkit-print-color-adjust: exact; print-color-adjust: exact;"></span> AI特征显著</span>
                <span class="legend-item"><span class="legend-color" style="background:#e5e7eb; -webkit-print-color-adjust: exact; print-color-adjust: exact;"></span> 未标识部分</span>
            </div>
        </div>
    </div>
    
    <div class="content">${contentHTML}</div>
</body>
</html>`;
}

function escapeHTML(str) {
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;');
}
