import { useEffect, useState, useRef } from "react";
import * as storage from "../lib/storage.js";
import { extractTextFromFile } from "../lib/pdf-parser.js";
import { splitWithStructure } from "../lib/splitter.js";
import { processWithSegments } from "../lib/randomizer.js";
import { downloadReport } from "../lib/report.js";
import Logo from "./Logo";
import DownloadModal from "./DownloadModal";

export default function UploadPage({ onBack }) {
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [file, setFile] = useState(null);
  const [tasks, setTasks] = useState([]);
  const [dragActive, setDragActive] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [pendingTask, setPendingTask] = useState(null);
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

  function handleDownloadClick(task) {
    if (!task.resultSentences) {
      alert("报告数据不存在");
      return;
    }
    setPendingTask(task);
    setShowModal(true);
  }

  function handleConfirmDownload() {
    setShowModal(false);
    if (pendingTask) {
      downloadReport(pendingTask.filename, pendingTask.resultSentences);
      setPendingTask(null);
    }
  }

  function formatFileSize(bytes) {
    if (!bytes) return "-";
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(2) + " MB";
  }

  function handleDrag(e) {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  }

  function handleDrop(e) {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const f = e.dataTransfer.files[0];
      if (f.type === "application/pdf") {
        setFile(f);
      } else {
        setMessage("请上传 PDF 格式的文件");
      }
    }
  }

  function renderStatus(status) {
    if (status === "success") {
      return (
        <span className="badge badge-success">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.5"><polyline points="20 6 9 17 4 12" /></svg>
          已完成
        </span>
      );
    } else if (status === "failed") {
      return <span className="badge badge-error">失败</span>;
    } else {
      return (
        <span className="badge badge-loading">
          <span className="spinner-sm" />
          处理中
        </span>
      );
    }
  }

  return (
    <div className="upload-page-bg">
    <div className="upload-page">
      <nav className="topbar">
        <button className="back-btn" onClick={onBack}>
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="19" y1="12" x2="5" y2="12" /><polyline points="12 19 5 12 12 5" />
          </svg>
          返回首页
        </button>
        <Logo width={160} height={48} />
        <div style={{ width: 80 }} />
      </nav>

      {message && (
        <div className={`toast ${message.includes("失败") ? "toast-error" : "toast-success"}`}>
          {message}
          <button className="toast-close" onClick={() => setMessage("")}>×</button>
        </div>
      )}

      {/* Upload Area */}
      <section className="upload-section">
        <h2>上传论文</h2>
        <div
          className={`drop-zone ${dragActive ? "drop-zone-active" : ""} ${file ? "drop-zone-has-file" : ""}`}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
        >
          <input ref={fileInputRef} type="file" accept="application/pdf" onChange={(e) => setFile(e.target.files[0])} style={{ display: "none" }} />
          {file ? (
            <div className="drop-zone-file">
              <svg viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="#f28219" strokeWidth="1.5">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" /><polyline points="14 2 14 8 20 8" /><line x1="16" y1="13" x2="8" y2="13" /><line x1="16" y1="17" x2="8" y2="17" /><polyline points="10 9 9 9 8 9" />
              </svg>
              <span className="drop-zone-filename">{file.name}</span>
              <span className="drop-zone-size">{formatFileSize(file.size)}</span>
            </div>
          ) : (
            <div className="drop-zone-empty">
              <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="#ccc" strokeWidth="1.5">
                <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" /><polyline points="17 8 12 3 7 8" /><line x1="12" y1="3" x2="12" y2="15" />
              </svg>
              <p>拖拽 PDF 文件到此处，或<span className="drop-zone-link">点击选择</span></p>
              <span className="drop-zone-hint">仅支持 PDF 格式</span>
            </div>
          )}
        </div>
        <button className="primary-btn" onClick={handleUpload} disabled={loading || !file}>
          {loading ? (
            <><span className="spinner-sm spinner-white" /> 检测中...</>
          ) : (
            <>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" /></svg>
              开始检测
            </>
          )}
        </button>
      </section>

      {/* Task List */}
      <section className="task-section">
        <h2>检测记录</h2>
        {tasks.length === 0 ? (
          <div className="empty-state">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="#ddd" strokeWidth="1">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" /><line x1="3" y1="9" x2="21" y2="9" /><line x1="9" y1="21" x2="9" y2="9" />
            </svg>
            <p>暂无检测记录</p>
          </div>
        ) : (
          <div className="task-cards">
            {tasks.map((task) => (
              <div className="task-card" key={task.taskId}>
                <div className="task-card-left">
                  <div className="task-card-name">{task.filename}</div>
                  <div className="task-card-meta">
                    <span>{formatFileSize(task.fileSize)}</span>
                    <span>·</span>
                    <span>{new Date(task.createdAt).toLocaleString("zh-CN")}</span>
                  </div>
                </div>
                <div className="task-card-right">
                  {renderStatus(task.status)}
                  {task.status === "success" && (
                    <button className="download-btn" onClick={() => handleDownloadClick(task)}>
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" /><polyline points="7 10 12 15 17 10" /><line x1="12" y1="15" x2="12" y2="3" /></svg>
                      下载报告
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {showModal && (
        <DownloadModal
          onConfirm={handleConfirmDownload}
          onCancel={() => { setShowModal(false); setPendingTask(null); }}
        />
      )}
    </div>
    </div>
  );
}
