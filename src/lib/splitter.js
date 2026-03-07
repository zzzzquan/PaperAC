/**
 * 文本分句器 —— 从 Go splitter 包移植并经过深度优化
 * 负责将 PDF 提取的文本切分为带类型的分段
 */

export const SegmentType = {
    BODY: 'body',
    TITLE: 'title',
    TABLE: 'table',
    LIST: 'list',
    META: 'meta',
    NEWLINE: 'newline',
    REFERENCE: 'reference', // 新增：参考文献段落
    CODE: 'code',           // 新增：代码或公式段落
};

const sentenceEndRe = /([。？！；!?;]+)/g;

// 优化：更严格的标题正则。强制要求是在句首的序号或核心关键词
const titlePatternRe = /^(\d+(\.\d+)+\s*)|^(第[一二三四五六七八九十百千]+[章节部分篇]\s*)|^(摘\s*要|Abstract|引\s*言|结\s*论|参考文献|致\s*谢|附\s*录|关键词|Keywords)$/i;

const listPatternRe = /^(\d+[.、)）]|\([0-9a-zA-Z]+\)|[•●○■□★☆\-–—]|[a-zA-Z][.、)）])/;
const tablePatternRe = /\s{2,}|\t/g;

// 优化：增加图表说明文字的正则拦截
const figureCaptionRe = /^(图|表|Fig\.|Table)\s*\d+/i;

const specialCharRe = /[#$%&*+/<=>@\\^_`{|}~\[\]\(\)\-×÷∑∏∫∂√∞≈≠≤≥±°′″αβγδεζηθικλμνξπρστυφχψω·]/;
const coverInfoRe = /(学号|姓名|班级|专业|学院|书院|题目|课程|论文|评阅|成绩|指导|教师|序号|负责|组别|全员|全英文|专题[一二三四五六七八九十]|\d{10,})/;
const pureNumberRe = /^\d+$|1[12]\d{8,}/;

// 新增：代码与数学公式正则
// 判断标准：大量英文字符连续、大量数学符号，或常见的代码保留关键字
const codeOrMathRe = /(\b(public|private|void|int|function|const|let|var|import|export|class|return\s+)\b)|([=+\-*/<>^~|&]{3,})/;

export function splitWithStructure(text) {
    const segments = [];
    const lines = text.split('\n');

    // 状态机变量
    let isReferenceSection = false; // 是否已进入参考文献区域
    let previousSegmentType = SegmentType.NEWLINE; // 记录上一行的类型

    for (let i = 0; i < lines.length; i++) {
        const trimmed = lines[i].trim();

        if (trimmed !== '') {

            // 1. 参考文献区域阻断
            // 一旦检测到参考文献主标题，后续所有内容都标记为 REFERENCE，不参与评分
            if (isReferenceSection) {
                segments.push({ text: trimmed, type: SegmentType.REFERENCE });
                continue;
            }

            // 检测类型需要传入上下文状态（行号、前一行类型）
            let segType = detectSegmentType(trimmed, i, previousSegmentType);

            // 检查是否刚刚进入了参考文献章节
            if (segType === SegmentType.TITLE && /^(参考文献|References)$/i.test(trimmed)) {
                isReferenceSection = true;
            }

            if (segType === SegmentType.BODY) {
                const sentences = splitSentences(trimmed);
                for (const s of sentences) {
                    const st = s.trim();
                    if (st === '') continue;

                    // shouldFilter 也要传入行号 i
                    if (shouldFilter(st, i)) {
                        segments.push({ text: st, type: SegmentType.META });
                    } else {
                        segments.push({ text: st, type: SegmentType.BODY });
                    }
                }
            } else {
                segments.push({ text: trimmed, type: segType });
            }

            previousSegmentType = segType;
        } else {
            previousSegmentType = SegmentType.NEWLINE;
        }

        if (i < lines.length - 1) {
            segments.push({ text: '', type: SegmentType.NEWLINE });
        }
    }

    return segments;
}

/**
 * 检查句子是否应该被过滤掉（归类为 meta）
 */
function shouldFilter(s, lineIndex) {
    if (specialCharRe.test(s)) return true;
    if (pureNumberRe.test(s)) return true;

    // 优化：封面信息仅在文档前 50 行内生效，避免误杀正文
    if (lineIndex < 50 && coverInfoRe.test(s)) {
        return true;
    }
    return false;
}

/**
 * 检测段落类型 (上下文感知)
 */
function detectSegmentType(para, lineIndex, previousType) {
    if (codeOrMathRe.test(para)) return SegmentType.CODE;
    if (listPatternRe.test(para)) return SegmentType.LIST;
    if (figureCaptionRe.test(para)) return SegmentType.META; // 图表说明

    // 优化：增强表格识别（上下文感知）
    if (isTableRow(para, previousType)) return SegmentType.TABLE;

    if (isTitle(para)) return SegmentType.TITLE;

    return SegmentType.BODY;
}

/**
 * 判断是否是标题
 */
function isTitle(para) {
    const runeCount = [...para].length;
    if (runeCount > 50) return false;

    if (titlePatternRe.test(para)) return true;

    if (para.endsWith('：') || para.endsWith(':')) return false;

    // 优化：普通短句被判定为标题的条件更加苛刻
    if (runeCount <= 25 && !sentenceEndRe.test(para)) {
        sentenceEndRe.lastIndex = 0;
        const connectors = ['的', '和', '与', '或', '但', '而', '是', '在', '了', '有', '为', '对', '从', '到'];
        for (const c of connectors) {
            if (para.endsWith(c)) return false;
        }
        let chineseCount = 0;
        for (const ch of para) {
            if (isChinese(ch)) chineseCount++;
        }
        // 如果汉字占比超过 80%（之前是 50%），才考虑是标题
        if (chineseCount / runeCount > 0.8) return true;
    }

    sentenceEndRe.lastIndex = 0;
    return false;
}

/**
 * 判断是否是表格行 (引入上下文状态)
 */
function isTableRow(para, previousType) {
    // 1. 如果有多个连续空格，标准表格行
    const matches = [...para.matchAll(new RegExp(tablePatternRe.source, 'g'))];
    if (matches.length >= 2) {
        const parts = para.split(new RegExp(tablePatternRe.source));
        let shortParts = 0;
        for (const p of parts) {
            const trimmed = p.trim();
            if (trimmed !== '' && [...trimmed].length <= 20) {
                shortParts++;
            }
        }
        if (shortParts >= 3) return true;
    }

    // 2. 优化：如果上一行是表格，当前行很短且没有句号结尾，也强制认为是同一表格的内容
    const runeCount = [...para].length;
    if (previousType === SegmentType.TABLE && runeCount < 40) {
        // 重置正则状态
        sentenceEndRe.lastIndex = 0;
        if (!sentenceEndRe.test(para)) {
            return true;
        }
        sentenceEndRe.lastIndex = 0;
    }

    return false;
}

/**
 * 将段落分割成句子
 */
function splitSentences(para) {
    const temp = para.replace(/([。？！；!?;]+)/g, '$1\n');
    return temp.split('\n').map(s => s.trim()).filter(s => s !== '');
}

/**
 * 判断汉字
 */
export function isChinese(ch) {
    const code = ch.codePointAt(0);
    return (code >= 0x4E00 && code <= 0x9FFF) ||
        (code >= 0x3400 && code <= 0x4DBF) ||
        (code >= 0x20000 && code <= 0x2A6DF) ||
        (code >= 0x2A700 && code <= 0x2B73F) ||
        (code >= 0x2B740 && code <= 0x2B81F) ||
        (code >= 0x2B820 && code <= 0x2CEAF) ||
        (code >= 0xF900 && code <= 0xFAFF) ||
        (code >= 0x2F800 && code <= 0x2FA1F);
}
