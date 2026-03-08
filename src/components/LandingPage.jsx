import Logo from "./Logo";

export default function LandingPage({ onStart }) {
  return (
    <div className="landing">
      <a
        className="github-corner"
        href="https://github.com/zzzzquan/PaperAC"
        target="_blank"
        rel="noopener noreferrer"
        title="GitHub"
      >
        <svg viewBox="0 0 24 24" width="28" height="28" fill="currentColor">
          <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
        </svg>
      </a>

      <section className="landing-hero">
        <div className="landing-hero-inner landing-reveal" style={{ animationDelay: "0.05s" }}>
          <div className="landing-logo">
            <Logo width={420} height={126} />
          </div>
        </div>
      </section>

      <section className="landing-panel">
        <div className="landing-curve" aria-hidden="true">
          <svg viewBox="0 0 100 100" preserveAspectRatio="none">
            <ellipse cx="50" cy="0" rx="55" ry="100" fill="#ffffff" />
          </svg>
        </div>
        <div className="landing-panel-inner landing-reveal" style={{ animationDelay: "0.16s" }}>
          <div className="feature-grid" role="list" aria-label="首页特性">
            <div className="feature-item" role="listitem">
              <div className="feature-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 12 20 22 4 22 4 12" />
                  <rect width="20" height="5" x="2" y="7" />
                  <line x1="12" x2="12" y1="22" y2="7" />
                  <path d="M12 7H7.5a2.5 2.5 0 0 1 0-5C11 2 12 7 12 7z" />
                  <path d="M12 7h4.5a2.5 2.5 0 0 0 0-5C13 2 12 7 12 7z" />
                </svg>
              </div>
              <h3>完全免费</h3>
            </div>

            <div className="feature-item" role="listitem">
              <div className="feature-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="10" />
                  <path d="M8 14s1.5 2 4 2 4-2 4-2" />
                  <line x1="9" x2="9.01" y1="9" y2="9" />
                  <line x1="15" x2="15.01" y1="9" y2="9" />
                </svg>
              </div>
              <h3>高情商</h3>
            </div>

            <div className="feature-item" role="listitem">
              <div className="feature-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" />
                  <path d="M9 12l2 2 4-4" />
                </svg>
              </div>
              <h3>完全隐私</h3>
            </div>
          </div>

          <div className="cta landing-reveal" style={{ animationDelay: "0.24s" }}>
            <button className="cta-btn" onClick={onStart}>
              立即体验
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="20" height="20">
                <line x1="5" y1="12" x2="19" y2="12" /><polyline points="12 5 19 12 12 19" />
              </svg>
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
