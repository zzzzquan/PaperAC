import * as pdfjsLib from 'pdfjs-dist';

// 设置 worker
pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
    'pdfjs-dist/build/pdf.worker.mjs',
    import.meta.url
).toString();

const sentenceEndPunctuationRe = /[。？！；.?!;]$/;
const chapterTitleRe = /^(第[一二三四五六七八九十百千\d]+[章节部分篇]|摘\s*要|Abstract|引\s*言|结\s*论|参考文献|致\s*谢|附\s*录)/;
const numericLineRe = /^\d+$/;

export async function extractTextFromFile(file) {
    const arrayBuffer = await file.arrayBuffer();
    const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;

    let fullText = '';
    for (let i = 1; i <= pdf.numPages; i++) {
        const page = await pdf.getPage(i);
        const content = await page.getTextContent();
        const pageText = content.items.map(item => item.str).join('');
        fullText += pageText + '\n';
    }

    const mergedText = preprocessLines(fullText);
    let normalized = normalizeText(mergedText);
    normalized = normalized.replace(/　/g, '');

    return normalized;
}

function preprocessLines(text) {
    const lines = text.split('\n');
    if (lines.length === 0) return '';

    const result = [];

    for (let i = 0; i < lines.length; i++) {
        const trimmed = lines[i].trim();

        if (trimmed === '') {
            result.push(lines[i]);
            continue;
        }

        let merged = trimmed;
        let nextIdx = i + 1;

        while (nextIdx < lines.length) {
            const nextTrimmed = lines[nextIdx].trim();
            if (nextTrimmed === '') {
                nextIdx++;
                continue;
            }

            let mergedUpdated = false;

            if (isShortChineseLine(merged)) {
                if (isShortChineseLine(nextTrimmed) || ([...nextTrimmed].length <= 4 && startsWithChinese(nextTrimmed))) {
                    merged += nextTrimmed;
                    mergedUpdated = true;
                }
            }

            if (!mergedUpdated && numericLineRe.test(nextTrimmed)) {
                merged += ' ' + nextTrimmed;
                mergedUpdated = true;
            }

            if (mergedUpdated) {
                i = nextIdx;
                nextIdx++;
            } else {
                break;
            }
        }
        result.push(merged);
    }
    return result.join('\n');
}

function isShortChineseLine(s) {
    const runes = [...s];
    if (runes.length === 0 || runes.length > 2) return false;
    return runes.some(ch => isChinese(ch));
}

function startsWithChinese(s) {
    const runes = [...s];
    if (runes.length === 0) return false;
    return isChinese(runes[0]);
}

function isChinese(ch) {
    const code = ch.codePointAt(0);
    return (code >= 0x4E00 && code <= 0x9FFF) ||
        (code >= 0x3400 && code <= 0x4DBF) ||
        (code >= 0x20000 && code <= 0x2A6DF);
}

function normalizeText(text) {
    const lines = text.split('\n');
    if (lines.length === 0) return '';

    const result = [];
    let currentParagraph = '';

    for (let i = 0; i < lines.length; i++) {
        const trimmed = lines[i].trim();

        if (trimmed === '') {
            if (currentParagraph.length > 0) {
                result.push(currentParagraph);
                currentParagraph = '';
            }
            continue;
        }

        if (isLikelyTitle(trimmed)) {
            if (currentParagraph.length > 0) {
                result.push(currentParagraph);
                currentParagraph = '';
            }
            result.push(trimmed);
            continue;
        }

        if (currentParagraph.length > 0) {
            if (sentenceEndPunctuationRe.test(currentParagraph)) {
                result.push(currentParagraph);
                currentParagraph = trimmed;
            } else {
                const space = needsSpaceBetween(currentParagraph, trimmed) ? ' ' : '';
                currentParagraph += space + trimmed;
            }
        } else {
            currentParagraph = trimmed;
        }

        if (i === lines.length - 1 && currentParagraph.length > 0) {
            result.push(currentParagraph);
        }
    }

    return result.join('\n');
}

function isLikelyTitle(line) {
    if ([...line].length > 50) return false;
    return chapterTitleRe.test(line);
}

function needsSpaceBetween(prev, next) {
    if (prev.length === 0 || next.length === 0) return false;
    const lastChar = prev[prev.length - 1];
    const firstChar = next[0];
    const prevIsLatin = /[a-zA-Z0-9]/.test(lastChar);
    const nextIsLatin = /[a-zA-Z0-9]/.test(firstChar);
    return prevIsLatin && nextIsLatin;
}
