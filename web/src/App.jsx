import { useEffect, useState } from "react";
import { fetchMe, logout, sendCode, verifyCode } from "./api";

export default function App() {
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [me, setMe] = useState(null);
  const [message, setMessage] = useState("");

  useEffect(() => {
    loadMe();
  }, []);

  async function loadMe() {
    try {
      const res = await fetchMe();
      setMe(res.data);
    } catch (err) {
      setMe(null);
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
      setMessage("验证码已发送，请查收邮件");
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
      setMessage("已退出登录");
    } catch (err) {
      setMessage(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page">
      <header className="header">
        <h1>PaperAC 认证入口</h1>
        <p>邮箱验证码登录（注册与登录合一）</p>
      </header>

      <section className="card">
        <h2>登录</h2>
        <label>
          邮箱
          <input
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
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
            onChange={(event) => setCode(event.target.value)}
            placeholder="6 位数字"
          />
        </label>
        <div className="actions">
          <button type="button" onClick={handleVerify} disabled={loading}>
            登录
          </button>
        </div>
        {message && <p className="message">{message}</p>}
      </section>

      <section className="card">
        <h2>当前用户</h2>
        {me ? (
          <div className="user">
            <p>用户ID: {me.user_id}</p>
            <p>邮箱: {me.email}</p>
            <button type="button" onClick={handleLogout} disabled={loading}>
              退出登录
            </button>
          </div>
        ) : (
          <p>未登录</p>
        )}
      </section>
    </div>
  );
}
