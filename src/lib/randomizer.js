import { SegmentType, isChinese } from './splitter.js';

const MIN_SENTENCE_CHARS = 20;

export function processWithSegments(segments) {
    if (!segments || segments.length === 0) {
        return { sentences: [], targetRatio: 0, actualRatio: 0 };
    }

    const targetRatio = (70 + Math.floor(Math.random() * 81)) / 10.0;
    const candidates = [];
    let totalBodyChars = 0;

    for (let i = 0; i < segments.length; i++) {
        const seg = segments[i];
        if (seg.type === SegmentType.BODY) {
            const charCount = [...seg.text].length;
            totalBodyChars += charCount;
            const validCount = countChineseChars(seg.text);
            if (validCount >= MIN_SENTENCE_CHARS) {
                candidates.push({ index: i, charCount });
            }
        }
    }

    const targetChars = Math.floor(totalBodyChars * targetRatio / 100.0);
    const aiIndices = selectDistributedSentences(candidates, targetChars);

    let actualAIChars = 0;
    const aiSet = new Set(aiIndices);
    for (const idx of aiIndices) {
        actualAIChars += [...segments[idx].text].length;
    }

    const actualRatio = totalBodyChars > 0 ? (actualAIChars / totalBodyChars) * 100.0 : 0;

    const sentences = segments.map((seg, i) => {
        let score, label;
        const segType = seg.type;

        if (seg.type !== SegmentType.BODY) {
            score = 0;
            label = 'structural';
        } else if (aiSet.has(i)) {
            score = 0.8 + Math.random() * 0.19;
            label = 'ai';
        } else {
            score = Math.random() * 0.2;
            label = 'human';
        }
        return { text: seg.text, score, label, type: segType };
    });

    return { sentences, targetRatio, actualRatio };
}

function countChineseChars(s) {
    let count = 0;
    for (const ch of s) {
        if (isChinese(ch)) count++;
    }
    return count;
}

function selectDistributedSentences(candidates, targetChars) {
    if (candidates.length === 0 || targetChars <= 0) return [];

    candidates.sort((a, b) => a.index - b.index);

    const used = new Set();
    const selected = [];
    let currentChars = 0;

    const blockRatio = 0.8;
    const blockQuota = Math.floor(targetChars * blockRatio);
    const minBlockSize = 3;
    const maxBlockSize = 6;
    const maxRetries = candidates.length * 2;
    let retries = 0;

    while (currentChars < blockQuota && retries < maxRetries) {
        retries++;
        const startIdx = Math.floor(Math.random() * candidates.length);
        const c = candidates[startIdx];
        if (used.has(c.index)) continue;

        const blockSize = minBlockSize + Math.floor(Math.random() * (maxBlockSize - minBlockSize + 1));
        const block = [];
        let blockChars = 0;
        let currSliceIdx = startIdx;
        let validBlock = true;

        for (let k = 0; k < blockSize; k++) {
            if (currSliceIdx >= candidates.length) break;
            const currCandidate = candidates[currSliceIdx];
            if (used.has(currCandidate.index)) break;
            if (k > 0) {
                const prevCandidate = candidates[currSliceIdx - 1];
                if (currCandidate.index !== prevCandidate.index + 1) break;
            }
            if (currentChars + blockChars + currCandidate.charCount > targetChars) {
                validBlock = false;
                break;
            }
            block.push(currCandidate.index);
            blockChars += currCandidate.charCount;
            currSliceIdx++;
        }

        if (block.length >= minBlockSize || (block.length > 0 && blockQuota - currentChars < 200)) {
            if (validBlock) {
                for (const idx of block) {
                    used.add(idx);
                    selected.push(idx);
                }
                currentChars += blockChars;
            }
        }
    }

    const perm = shuffledIndices(candidates.length);
    for (const i of perm) {
        if (currentChars >= targetChars) break;
        const c = candidates[i];
        if (!used.has(c.index)) {
            if (currentChars + c.charCount <= targetChars + 100) {
                used.add(c.index);
                selected.push(c.index);
                currentChars += c.charCount;
            }
        }
    }

    return selected;
}

function shuffledIndices(n) {
    const arr = Array.from({ length: n }, (_, i) => i);
    for (let i = arr.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [arr[i], arr[j]] = [arr[j], arr[i]];
    }
    return arr;
}
