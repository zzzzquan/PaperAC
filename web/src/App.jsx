import { useEffect, useState } from "react";
import { createTask, fetchTasks, previewTask, clearSession } from "./api";

export default function App() {
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");

  // 任务相关状态
  const [file, setFile] = useState(null);
  const [tasks, setTasks] = useState([]);

  // 自动刷新任务列表
  useEffect(() => {
    loadTasks();
    const timer = setInterval(loadTasks, 3000);
    return () => clearInterval(timer);
  }, []);

  // 页面关闭/刷新时尝试清理会话
  // 注意：刷新也会触发清理，这意味着刷新后会话丢失，符合"Exit"即退出的隐私要求
  useEffect(() => {
    const handleUnload = () => {
      // 使用 fetch keepalive 尝试清理
      // 我们不能直接调用 clearSession 因为它是异步的且没用 keepalive
      // 这里手动构建请求
      const sid = sessionStorage.getItem("paperac_sid");
      if (sid) {
        fetch("/api/session", {
          method: "DELETE",
          headers: { "X-Session-ID": sid },
          keepalive: true
        });
      }
    };
    window.addEventListener("beforeunload", handleUnload);
    return () => window.removeEventListener("beforeunload", handleUnload);
  }, []);


  async function loadTasks() {
    try {
      const res = await fetchTasks();
      setTasks(res.data.items || []);
    } catch (err) {
      console.error(err);
    }
  }

  async function handleUpload() {
    if (!file) {
      setMessage("请选择PDF文件");
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      // Hardcode x=0.7 as per requirement to clean up UI input
      await createTask(file, 0.7);
      setMessage("任务创建成功");
      setFile(null);
      // 重置文件输入
      const fileInput = document.querySelector('input[type="file"]');
      if (fileInput) fileInput.value = "";
      loadTasks(); // 立即刷新
    } catch (err) {
      setMessage(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handlePreview(task) {
    try {
      await previewTask(task.task_id);
    } catch (err) {
      alert("预览失败: " + err.message);
    }
  }

  async function handleClearSession() {
    if (!confirm("确定要结束会话并清除所有记录吗？")) return;
    try {
      await clearSession();
      setTasks([]);
      setMessage("会话已结束，记录已清除");
    } catch (err) {
      alert("清理失败: " + err.message);
    }
  }

  function formatFileSize(bytes) {
    if (!bytes) return "-";
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(2) + " MB";
  }

  function renderStatus(status) {
    if (status === 'success') {
      return (
        <div className="status-icon checkmark" title="已完成">
          <svg viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" /></svg>
        </div>
      );
    } else if (status === 'failed') {
      return (
        <div className="status-icon failed-icon" title="失败">✖</div>
      );
    } else {
      // pending, processing
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
        <div className="header-left">
          <h1>PaperAC 文鉴</h1>
          <p>高情商论文AIGC检测工具</p>
        </div>
        <button className="link-btn" style={{ color: '#999' }} onClick={handleClearSession}>
          结束会话
        </button>
      </header>

      {message && <p className="message info">{message}</p>}

      <section className="card">
        <h2>创建新任务</h2>
        <div className="form-group">
          <label>上传PDF论文</label>
          <input
            type="file"
            accept="application/pdf"
            onChange={(e) => setFile(e.target.files[0])}
          />
        </div>
        <div className="actions">
          <button onClick={handleUpload} disabled={loading || !file}>
            开始检测
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
                {tasks.map(task => (
                  <tr key={task.task_id}>
                    <td>{task.filename}</td>
                    <td>{formatFileSize(task.file_size)}</td>
                    <td>{new Date(task.created_at).toLocaleString()}</td>
                    <td>
                      {renderStatus(task.status)}
                    </td>
                    <td>
                      {task.status === 'success' ? (
                        <button
                          className="link-btn"
                          onClick={() => handlePreview(task)}
                        >
                          查看报告
                        </button>
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

