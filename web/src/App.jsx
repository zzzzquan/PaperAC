import { useEffect, useState } from "react";
import { fetchMe, logout, sendCode, verifyCode, createTask, fetchTasks } from "./api";

export default function App() {
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [me, setMe] = useState(null);
  const [message, setMessage] = useState("");

  // 任务相关状态
  const [file, setFile] = useState(null);
  const [xVal, setXVal] = useState(0.7);
  const [tasks, setTasks] = useState([]);

  useEffect(() => {
    loadMe();
  }, []);

  // 自动刷新任务列表
  useEffect(() => {
    if (!me) return;
    loadTasks();
    const timer = setInterval(loadTasks, 3000);
    return () => clearInterval(timer);
  }, [me]);

  async function loadMe() {
    try {
      const res = await fetchMe();
      setMe(res.data);
    } catch (err) {
      setMe(null);
    }
  }

  async function loadTasks() {
    try {
      const res = await fetchTasks();
      setTasks(res.data.items || []);
    } catch (err) {
      console.error(err);
    }
  }

  async function handleSendCode() {
    if (!email) {
      setMessage("请输入邮箱");
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      await sendCode(email);
      setMessage("验证码已发送（测试模式请查看日志）");
    } catch (err) {
      setMessage(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleVerify() {
    if (!email || !code) {
      setMessage("请输入邮箱与验证码");
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      const res = await verifyCode(email, code);
      setMe(res.data);
      setMessage("登录成功");
    } catch (err) {
      setMessage(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleLogout() {
    setLoading(true);
    setMessage("");
    try {
      await logout();
      setMe(null);
      setTasks([]);
      setMessage("已退出登录");
    } catch (err) {
      setMessage(err.message);
    } finally {
      setLoading(false);
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
      await createTask(file, xVal);
      setMessage("任务创建成功");
      setFile(null);
      loadTasks(); // 立即刷新
    } catch (err) {
      setMessage(err.message);
    } finally {
      setLoading(false);
    }
  }

  // 登录前的页面
  if (!me) {
    return (
      <div className="page">
        <header className="header">
          <h1>PaperAC 文鉴</h1>
          <p>高情商论文AIGC检测工具</p>
        </header>

        <section className="card">
          <h2>登录</h2>
          <label>
            邮箱
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="user@example.com"
            />
          </label>
          <div className="actions">
            <button type="button" onClick={handleSendCode} disabled={loading}>
              发送验证码
            </button>
          </div>
          <label>
            验证码
            <input
              type="text"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="6 位数字"
            />
          </label>
          <div className="actions">
            <button type="button" onClick={handleVerify} disabled={loading}>
              登录
            </button>
          </div>
          {message && <p className="message error">{message}</p>}
        </section>
      </div>
    );
  }

  // 登录后的页面
  return (
    <div className="page">
      <header className="header">
        <h1>PaperAC 文鉴</h1>
        <div className="user-info">
          <span>{me.email}</span>
          <button className="small-btn" onClick={handleLogout} disabled={loading}>退出</button>
        </div>
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
        <div className="form-group">
          <label>AIGC 目标占比 (x)</label>
          <input
            type="number"
            min="0" max="1" step="0.1"
            value={xVal}
            onChange={(e) => setXVal(parseFloat(e.target.value))}
          />
          <span className="hint">范围 0 ~ 1 （例如 0.3 表示 30%）</span>
        </div>
        <div className="actions">
          <button onClick={handleUpload} disabled={loading || !file}>
            开始检测
          </button>
        </div>
      </section>

      <section className="card full-width">
        <h2>任务列表</h2>
        <div className="task-list">
          {tasks.length === 0 ? (
            <p className="empty">暂无任务</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>文件名</th>
                  <th>提交时间</th>
                  <th>状态</th>
                  <th>进度</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map(task => (
                  <tr key={task.task_id}>
                    <td>{task.filename}</td>
                    <td>{new Date(task.created_at).toLocaleString()}</td>
                    <td>
                      <span className={`status-badge ${task.status}`}>
                        {task.status}
                      </span>
                    </td>
                    <td>
                      <div className="progress-bar">
                        <div className="fill" style={{ width: `${task.progress}%` }}></div>
                      </div>
                      <span className="progress-text">{task.progress}%</span>
                    </td>
                    <td>
                      {task.status === 'success' ? (
                        <a
                          href={`/api/tasks/${task.task_id}/result`}
                          target="_blank"
                          rel="noreferrer"
                          className="download-link"
                        >
                          下载报告
                        </a>
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
