import { useEffect, useState, useRef } from "react";
import * as storage from "./lib/storage.js";
import { extractTextFromFile } from "./lib/pdf-parser.js";
import { splitWithStructure } from "./lib/splitter.js";
import { processWithSegments } from "./lib/randomizer.js";
import { previewReport } from "./lib/report.js";

export default function App() {
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [file, setFile] = useState(null);
  const [tasks, setTasks] = useState([]);
  const fileInputRef = useRef(null);

  useEffect(() => {
    setTasks(storage.listTasks());
  }, []);

  async function handleUpload() {
    if (!file) {
      setMessage("请选择PDF文件");
      return;
    }
    setLoading(true);
    setMessage("");

    const task = storage.createTask(file.name, file.size);
    refreshTasks();

    try {
      storage.updateTask(task.taskId, { status: "running", progress: 10 });
      refreshTasks();

      const text = await extractTextFromFile(file);
      storage.updateTask(task.taskId, { progress: 40 });
      refreshTasks();

      const segments = splitWithStructure(text);
      storage.updateTask(task.taskId, { progress: 60 });
      refreshTasks();

      const result = processWithSegments(segments);
      storage.updateTask(task.taskId, { progress: 80 });
      refreshTasks();

      storage.updateTask(task.taskId, {
        status: "success",
        progress: 100,
        resultSentences: result.sentences,
        finishedAt: new Date().toISOString(),
      });

      setMessage("检测完成！");
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
    } catch (err) {
      console.error("处理失败:", err);
      storage.updateTask(task.taskId, {
        status: "failed",
        errorMessage: err.message || "未知错误",
        finishedAt: new Date().toISOString(),
      });
      setMessage("处理失败: " + (err.message || "未知错误"));
    } finally {
      setLoading(false);
      refreshTasks();
    }
  }

  function refreshTasks() {
    setTasks(storage.listTasks());
  }

  function handlePreview(task) {
    if (!task.resultSentences) {
      alert("报告数据不存在");
      return;
    }
    previewReport(task.filename, task.resultSentences);
  }

  function handleClearSession() {
    if (!confirm("确定要结束会话并清除所有记录吗？")) return;
    storage.clearSession();
    setTasks([]);
    setMessage("会话已结束，记录已清除");
  }

  function formatFileSize(bytes) {
    if (!bytes) return "-";
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(2) + " MB";
  }

  function renderStatus(status) {
    if (status === "success") {
      return (
        <div className="status-icon checkmark" title="已完成">
          <svg viewBox="0 0 24 24">
            <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
          </svg>
        </div>
      );
    } else if (status === "failed") {
      return (
        <div className="status-icon failed-icon" title="失败">✖</div>
      );
    } else {
      return (
        <div className="status-icon">
          <div className="spinner" title="处理中..."></div>
        </div>
      );
    }
  }

  return (
    <div className="page">
      <header className="header">
        <div class="header-left">
          <h1>paperAC</h1>
          <p>高情商论文AIGC检测工具</p>
        </div>
        <button className="link-btn" style={{ color: "#999" }} onClick={handleClearSession}>
          结束会话
        </button>
      </header>

      {message && <p className="message info">{message}</p>}

      <section className="card">
        <h2>创建新任务</h2>
        <div className="form-group">
          <label>上传PDF论文</label>
          <input
            ref={fileInputRef}
            type="file"
            accept="application/pdf"
            onChange={(e) => setFile(e.target.files[0])}
          />
        </div>
        <div className="actions">
          <button onClick={handleUpload} disabled={loading || !file}>
            {loading ? "检测中..." : "开始检测"}
          </button>
        </div>
      </section>

      <section className="card full-width">
        <h2>最近任务</h2>
        <div className="task-list">
          {tasks.length === 0 ? (
            <p className="empty">暂无任务</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>文件名</th>
                  <th>大小</th>
                  <th>提交时间</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map((task) => (
                  <tr key={task.taskId}>
                    <td>{task.filename}</td>
                    <td>{formatFileSize(task.fileSize)}</td>
                    <td>{new Date(task.createdAt).toLocaleString()}</td>
                    <td>{renderStatus(task.status)}</td>
                    <td>
                      {task.status === "success" ? (
                        <button className="link-btn" onClick={() => handlePreview(task)}>查看报告</button>
                      ) : (
                        <span className="disabled-text">-</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </section>
    </div>
  );
}
