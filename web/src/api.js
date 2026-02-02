// API 函数（无认证版本）

export async function createTask(file, x) {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("x", x);

  return request("/api/tasks", {
    method: "POST",
    body: formData
  });
}

export async function fetchTasks(limit = 20) {
  return request(`/api/tasks?limit=${limit}`, { method: "GET" });
}

export async function previewTask(taskId) {
  const res = await fetch(`/api/tasks/${taskId}/result`, {
    method: "GET",
  });

  if (!res.ok) {
    throw new Error("预览失败: " + res.status);
  }

  const blob = await res.blob();
  const url = window.URL.createObjectURL(blob);
  window.open(url, "_blank");
}

async function request(path, options) {
  const headers = { ...options.headers };

  // Only set JSON content type if not using FormData
  if (!(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const res = await fetch(path, {
    ...options,
    headers: headers
  });

  const payload = await res.json().catch(() => null);
  if (!res.ok || (payload && payload.code !== 0)) {
    const message = payload?.message || "请求失败";
    const code = payload?.code ?? 9000;
    throw new Error(`${code}:${message}`);
  }
  return payload;
}
