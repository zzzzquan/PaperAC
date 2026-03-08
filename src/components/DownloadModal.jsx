import { useState } from "react";

export default function DownloadModal({ onConfirm, onCancel }) {
  const [agreed, setAgreed] = useState(false);

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-card" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <div className="modal-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="28" height="28">
              <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
          </div>
          <h3>下载报告</h3>
        </div>

        <div className="modal-body">
          <div className="modal-tip">
            <div className="tip-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="#15803d" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M9 18c0 1.1.9 2 2 2h2a2 2 0 002-2" />
                <path d="M12 2a6 6 0 014 10.5V15a1 1 0 01-1 1H9a1 1 0 01-1-1v-2.5A6 6 0 0112 2z" />
              </svg>
            </div>
            <div className="tip-text">
              点击弹出窗口左下角 <strong>「保存」</strong>，即可下载该报告。
            </div>
          </div>

          <div className="modal-warning">
            <div className="warning-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="#b45309" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </svg>
            </div>
            <div className="warning-text">
              本项目仅供<strong>娱乐</strong>，请勿用于真实AIGC检测场景。所发生的一切后果，本项目概不负责！
            </div>
          </div>
        </div>

        <div className="modal-footer">
          <label className="modal-checkbox">
            <input type="checkbox" checked={agreed} onChange={(e) => setAgreed(e.target.checked)} />
            <span>我已阅读并了解以上内容</span>
          </label>
          <div className="modal-actions">
            <button className="modal-btn-cancel" onClick={onCancel}>取消</button>
            <button className="modal-btn-confirm" disabled={!agreed} onClick={onConfirm}>
              确认下载
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
