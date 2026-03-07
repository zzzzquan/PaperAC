/**
 * 文本分句器 —— 从 Go splitter 包移植
 * 负责将 PDF 提取的文本切分为带类型的分段
 */

export const SegmentType = {
    BODY: 'body',
    TITLE: 'title',
    TABLE: 'table',
    LIST: 'list',
    META: 'meta',
    NEWLINE: 'newline',
};

const sentenceEndRe = /([。？！；!?;]+)/g;
const titlePatternRe = /^(第[一二三四五六七八九十百千\d]+[章节部分篇]|摘\s*要|Abstract|引\s*言|结\s*论|参考文献|致\s*谢|附\s*录|关键词|Keywords)/;
const listPatternRe = /^(\d+[.、)）]|\([0-9a-zA-Z]+\)|[•●○■□★☆\-–—]|[a-zA-Z][.、)）])/;
const tablePatternRe = /\s{2,}|\t/g;
const specialCharRe = /[#$%&*+/<=>@\\^_`{|}~\[\]\(\)\-×÷∑∏∫∂√∞≈≠≤≥±°′″αβγδεζηθικλμνξπρστυφχψω·]/;
const coverInfoRe = /(学号|姓名|班级|专业|学院|书院|题目|课程|论文|评阅|成绩|指导|教师|序号|负责|组别|全员|全英文|专题[一二三四五六七八九十]|\d{10,})/;
const pureNumberRe = /^\d+$|1[12]\d{8,}/;

export function splitWithStructure(text) {
    const segments = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
        const trimmed = lines[i].trim();

        if (trimmed !== '') {
            const segType = detectSegmentType(trimmed);
            if (segType === SegmentType.BODY) {
                const sentences = splitSentences(trimmed);
                for (const s of sentences) {
                    const st = s.trim();
                    if (st === '') continue;
                    if (shouldFilter(st)) {
                        segments.push({ text: st, type: SegmentType.META });
                    } else {
                        segments.push({ text: st, type: SegmentType.BODY });
                    }
                }
            } else {
                segments.push({ text: trimmed, type: segType });
            }
        }

        if (i < lines.length - 1) {
            segments.push({ text: '', type: SegmentType.NEWLINE });
        }
    }

    return segments;
}

function shouldFilter(s) {
    if (specialCharRe.test(s)) return true;
    if (coverInfoRe.test(s)) return true;
    if (pureNumberRe.test(s)) return true;
    return false;
}

function detectSegmentType(para) {
    if (listPatternRe.test(para)) return SegmentType.LIST;
    if (isTableRow(para)) return SegmentType.TABLE;
    if (isTitle(para)) return SegmentType.TITLE;
    return SegmentType.BODY;
}

function isTitle(para) {
    const runeCount = [...para].length;
    if (runeCount > 50) return false;
    if (titlePatternRe.test(para)) return true;
    if (para.endsWith('：') || para.endsWith(':')) return false;

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
        if (chineseCount / runeCount > 0.5) return true;
    }
    sentenceEndRe.lastIndex = 0;
    return false;
}

function isTableRow(para) {
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
        return shortParts >= 3;
    }
    return false;
}

function splitSentences(para) {
    const temp = para.replace(/([。？！；!?;]+)/g, '$1\n');
    return temp.split('\n').map(s => s.trim()).filter(s => s !== '');
}

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
