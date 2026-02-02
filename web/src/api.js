export async function sendCode(email) {
  return request("/api/auth/send-code", {
    method: "POST",
    body: JSON.stringify({ email })
  });
}

export async function verifyCode(email, code) {
  return request("/api/auth/verify", {
    method: "POST",
    body: JSON.stringify({ email, code })
  });
}

export async function fetchMe() {
  return request("/api/auth/me", { method: "GET" });
}

export async function logout() {
  return request("/api/auth/logout", {
    method: "POST",
    headers: {
      "X-CSRF-Token": getCSRFCookie()
    }
  });
}

export async function createTask(file, x) {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("x", x);

  // Note: Do not set Content-Type header when using FormData; 
  // fetch/browser will set it to multipart/form-data with boundary automatically.
  return request("/api/tasks", {
    method: "POST",
    headers: {
      "X-CSRF-Token": getCSRFCookie()
      // "Content-Type" is intentionally omitted
    },
    body: formData
  });
}

export async function fetchTasks(limit = 20) {
  return request(`/api/tasks?limit=${limit}`, { method: "GET" });
}

export async function previewTask(taskId) {
  const token = getToken();
  const headers = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`/api/tasks/${taskId}/result`, {
    method: "GET",
    headers: headers,
  });

  if (!res.ok) {
    throw new Error("预览失败: " + res.status);
  }

  const blob = await res.blob();
  const url = window.URL.createObjectURL(blob);
  window.open(url, "_blank");
}

// Token management
const TOKEN_KEY = 'auth_token';

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

async function request(path, options) {
  const headers = { ...options.headers };
  // Only set JSON content type if not using FormData (which shouldn't have content-type set manually)
  if (!(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  // Add Authorization header if token exists
  const token = getToken();
  if (token) {
    console.log("[DEBUG] Attaching Token:", token.substring(0, 10) + "...");
    headers["Authorization"] = `Bearer ${token}`;
  } else {
    console.log("[DEBUG] No Token found in localStorage");
  }

  const res = await fetch(path, {
    credentials: "include",
    ...options,
    headers: headers
  });

  const payload = await res.json().catch(() => null);
  if (!res.ok || (payload && payload.code !== 0)) {
    // If 401, maybe clear token? For now just throw.
    if (res.status === 401) {
      // removeToken(); // Optional: auto-logout on 401
    }
    const message = payload?.message || "请求失败";
    const code = payload?.code ?? 9000;
    throw new Error(`${code}:${message}`);
  }
  return payload;
}

function getCSRFCookie() {
  const match = document.cookie.match(/(^|;)\s*csrf_token=([^;]+)/);
  return match ? decodeURIComponent(match[2]) : "";
}
