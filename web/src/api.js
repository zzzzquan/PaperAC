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

async function request(path, options) {
  const res = await fetch(path, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {})
    },
    ...options
  });

  const payload = await res.json().catch(() => null);
  if (!res.ok || (payload && payload.code !== 0)) {
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
