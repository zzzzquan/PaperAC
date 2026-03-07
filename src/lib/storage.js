const TASKS_KEY = 'paperac_tasks';
const SESSION_KEY = 'paperac_sid';

export function getSessionID() {
    let sid = sessionStorage.getItem(SESSION_KEY);
    if (!sid) {
        sid = crypto.randomUUID();
        sessionStorage.setItem(SESSION_KEY, sid);
    }
    return sid;
}

export function listTasks() {
    const sid = getSessionID();
    const all = getAllTasks();
    return all
        .filter(t => t.sessionId === sid)
        .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
}

export function createTask(filename, fileSize) {
    const task = {
        taskId: crypto.randomUUID(),
        sessionId: getSessionID(),
        filename,
        fileSize,
        status: 'pending',
        progress: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        finishedAt: null,
        errorMessage: '',
        resultSentences: null,
    };

    const all = getAllTasks();
    all.push(task);
    saveAllTasks(all);
    return task;
}

export function updateTask(taskId, updates) {
    const all = getAllTasks();
    const idx = all.findIndex(t => t.taskId === taskId);
    if (idx === -1) return null;

    all[idx] = { ...all[idx], ...updates, updatedAt: new Date().toISOString() };
    saveAllTasks(all);
    return all[idx];
}

export function getTask(taskId) {
    const all = getAllTasks();
    return all.find(t => t.taskId === taskId) || null;
}

export function clearSession() {
    const sid = getSessionID();
    const all = getAllTasks();
    const filtered = all.filter(t => t.sessionId !== sid);
    saveAllTasks(filtered);
    sessionStorage.removeItem(SESSION_KEY);
}

function getAllTasks() {
    try {
        const raw = localStorage.getItem(TASKS_KEY);
        return raw ? JSON.parse(raw) : [];
    } catch {
        return [];
    }
}

function saveAllTasks(tasks) {
    localStorage.setItem(TASKS_KEY, JSON.stringify(tasks));
}
