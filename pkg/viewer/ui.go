package viewer

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) handleDefaultUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, getHTMLDashboard())
}

func getHTMLDashboard() string {
	return strings.ReplaceAll(dashboardTemplate, "__BT__", "`")
}

const dashboardTemplate = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Security Finding Dashboard — go-code-scanner</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-base: #080c14;
      --bg-surface: #0e1422;
      --bg-elevated: #141c2e;
      --bg-card: rgba(19, 28, 46, 0.7);
      --bg-card-hover: rgba(28, 40, 65, 0.85);
      --bg-input: #090e18;
      
      --border-subtle: rgba(255, 255, 255, 0.08);
      --border-medium: rgba(255, 255, 255, 0.14);
      --border-accent: rgba(99, 102, 241, 0.4);
      
      --text-primary: #f8fafc;
      --text-secondary: #94a3b8;
      --text-muted: #64748b;
      
      --accent-indigo: #6366f1;
      --accent-purple: #a855f7;
      --accent-cyan: #06b6d4;
      --accent-gradient: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
      --accent-gradient-hover: linear-gradient(135deg, #4f46e5 0%, #9333ea 100%);
      
      --sev-critical: #f43f5e;
      --sev-critical-bg: rgba(244, 63, 94, 0.12);
      --sev-critical-border: rgba(244, 63, 94, 0.35);
      
      --sev-high: #f97316;
      --sev-high-bg: rgba(249, 115, 22, 0.12);
      --sev-high-border: rgba(249, 115, 22, 0.35);
      
      --sev-medium: #f59e0b;
      --sev-medium-bg: rgba(245, 158, 11, 0.12);
      --sev-medium-border: rgba(245, 158, 11, 0.35);
      
      --sev-low: #38bdf8;
      --sev-low-bg: rgba(56, 189, 248, 0.12);
      --sev-low-border: rgba(56, 189, 248, 0.35);
      
      --sev-info: #94a3b8;
      --sev-info-bg: rgba(148, 163, 184, 0.12);
      --sev-info-border: rgba(148, 163, 184, 0.3);

      --success-green: #10b981;
      --success-bg: rgba(16, 185, 129, 0.12);
      
      --radius-sm: 6px;
      --radius-md: 10px;
      --radius-lg: 16px;
      --radius-xl: 20px;
      
      --shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.3);
      --shadow-md: 0 8px 24px rgba(0, 0, 0, 0.4);
      --shadow-lg: 0 16px 40px rgba(0, 0, 0, 0.6);
      --shadow-glow: 0 0 24px rgba(99, 102, 241, 0.3);
      --shadow-crit-glow: 0 0 20px rgba(244, 63, 94, 0.25);
    }

    * { box-sizing: border-box; margin: 0; padding: 0; }
    
    body {
      font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
      background-color: var(--bg-base);
      background-image: 
        radial-gradient(circle at 15% 50%, rgba(99, 102, 241, 0.06) 0%, transparent 40%),
        radial-gradient(circle at 85% 20%, rgba(168, 85, 247, 0.05) 0%, transparent 40%);
      color: var(--text-primary);
      min-height: 100vh;
      line-height: 1.5;
      -webkit-font-smoothing: antialiased;
    }

    .mono { font-family: 'JetBrains Mono', monospace; }

    /* Container */
    .app-container {
      max-width: 1440px;
      margin: 0 auto;
      padding: 24px 28px 60px;
    }

    /* Top Navigation Header */
    header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 22px;
      background: var(--bg-card);
      backdrop-filter: blur(20px);
      -webkit-backdrop-filter: blur(20px);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg);
      margin-bottom: 24px;
      box-shadow: var(--shadow-sm);
    }

    .brand-section {
      display: flex;
      align-items: center;
      gap: 14px;
    }

    .brand-logo {
      width: 42px;
      height: 42px;
      border-radius: var(--radius-md);
      background: var(--accent-gradient);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      box-shadow: var(--shadow-glow);
    }

    .brand-titles h1 {
      font-size: 1.08rem;
      font-weight: 700;
      letter-spacing: -0.02em;
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .brand-titles .version-tag {
      font-size: 0.68rem;
      font-weight: 600;
      background: rgba(99, 102, 241, 0.15);
      color: #a5b4fc;
      border: 1px solid rgba(99, 102, 241, 0.3);
      padding: 1px 7px;
      border-radius: 9999px;
    }

    .brand-titles p {
      font-size: 0.76rem;
      color: var(--text-secondary);
    }

    .header-controls {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .pill-info {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 5px 12px;
      border-radius: 9999px;
      font-size: 0.76rem;
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid var(--border-subtle);
      color: var(--text-secondary);
    }

    .status-indicator {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 5px 12px;
      border-radius: 9999px;
      font-size: 0.74rem;
      font-weight: 600;
      background: var(--success-bg);
      border: 1px solid rgba(16, 185, 129, 0.3);
      color: var(--success-green);
    }

    .pulse-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: var(--success-green);
      box-shadow: 0 0 8px var(--success-green);
      animation: pulse 2s infinite;
    }

    @keyframes pulse {
      0% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.4; transform: scale(0.85); }
      100% { opacity: 1; transform: scale(1); }
    }

    /* Executive Security Banner */
    .executive-grid {
      display: grid;
      grid-template-columns: 320px 1fr;
      gap: 20px;
      margin-bottom: 24px;
    }

    @media (max-width: 1024px) {
      .executive-grid { grid-template-columns: 1fr; }
    }

    .score-card {
      background: var(--bg-card);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg);
      padding: 24px;
      display: flex;
      align-items: center;
      gap: 20px;
      position: relative;
      overflow: hidden;
      box-shadow: var(--shadow-sm);
    }

    .score-gauge {
      width: 90px;
      height: 90px;
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .score-gauge svg {
      transform: rotate(-90deg);
      width: 90px;
      height: 90px;
    }

    .score-gauge circle {
      fill: none;
      stroke-width: 8;
      stroke-linecap: round;
    }

    .score-gauge .bg-circle { stroke: rgba(255, 255, 255, 0.08); }
    .score-gauge .prog-circle {
      stroke: var(--success-green);
      stroke-dasharray: 238;
      stroke-dashoffset: 0;
      transition: stroke-dashoffset 0.8s ease, stroke 0.4s ease;
    }

    .score-label {
      position: absolute;
      text-align: center;
    }

    .score-number {
      font-size: 1.5rem;
      font-weight: 800;
      line-height: 1;
    }

    .score-grade {
      font-size: 0.7rem;
      font-weight: 700;
      color: var(--text-muted);
    }

    .score-details h3 {
      font-size: 1.05rem;
      font-weight: 700;
      letter-spacing: -0.01em;
      margin-bottom: 4px;
    }

    .score-details p {
      font-size: 0.78rem;
      color: var(--text-secondary);
      line-height: 1.4;
    }

    .policy-tag {
      display: inline-flex;
      align-items: center;
      gap: 5px;
      margin-top: 8px;
      font-size: 0.72rem;
      font-weight: 700;
      padding: 3px 9px;
      border-radius: 4px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .policy-tag.passed { background: var(--success-bg); color: var(--success-green); border: 1px solid rgba(16,185,129,0.3); }
    .policy-tag.failed { background: var(--sev-critical-bg); color: var(--sev-critical); border: 1px solid var(--sev-critical-border); }

    /* KPI Metrics Cards */
    .kpi-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 16px;
    }

    .kpi-card {
      background: var(--bg-card);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg);
      padding: 18px 20px;
      position: relative;
      transition: transform 0.2s ease, border-color 0.2s ease;
      box-shadow: var(--shadow-sm);
    }

    .kpi-card:hover {
      transform: translateY(-2px);
      border-color: var(--border-medium);
    }

    .kpi-title {
      font-size: 0.72rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--text-muted);
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .kpi-value {
      font-size: 1.85rem;
      font-weight: 800;
      margin: 6px 0 2px;
      letter-spacing: -0.03em;
    }

    .kpi-sub {
      font-size: 0.72rem;
      color: var(--text-secondary);
    }

    .kpi-card.critical .kpi-value { color: var(--sev-critical); }
    .kpi-card.high .kpi-value { color: var(--sev-high); }
    .kpi-card.medium .kpi-value { color: var(--sev-medium); }
    .kpi-card.low .kpi-value { color: var(--sev-low); }

    /* Scan Action Bar */
    .scan-action-bar {
      background: var(--bg-surface);
      border: 1px solid var(--border-medium);
      border-radius: var(--radius-lg);
      padding: 16px 20px;
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 14px;
      margin-bottom: 24px;
      box-shadow: var(--shadow-md);
    }

    .form-item {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 0.82rem;
      color: var(--text-secondary);
      font-weight: 500;
    }

    select, input[type="text"] {
      background: var(--bg-input);
      border: 1px solid var(--border-subtle);
      color: var(--text-primary);
      padding: 8px 12px;
      border-radius: var(--radius-sm);
      font-family: inherit;
      font-size: 0.82rem;
      font-weight: 500;
      outline: none;
      transition: all 0.2s ease;
    }

    select:focus, input[type="text"]:focus {
      border-color: var(--accent-indigo);
      box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
    }

    .custom-checkbox {
      display: flex;
      align-items: center;
      gap: 7px;
      font-size: 0.82rem;
      font-weight: 500;
      color: var(--text-secondary);
      cursor: pointer;
      user-select: none;
    }

    .custom-checkbox input {
      accent-color: var(--accent-indigo);
      width: 15px;
      height: 15px;
      cursor: pointer;
    }

    .btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      padding: 9px 18px;
      border-radius: var(--radius-sm);
      font-size: 0.82rem;
      font-weight: 600;
      cursor: pointer;
      border: none;
      transition: all 0.2s ease;
      font-family: inherit;
      white-space: nowrap;
    }

    .btn-primary {
      background: var(--accent-gradient);
      color: #fff;
      box-shadow: 0 4px 14px rgba(99, 102, 241, 0.35);
    }

    .btn-primary:hover:not(:disabled) {
      background: var(--accent-gradient-hover);
      transform: translateY(-1px);
      box-shadow: 0 6px 18px rgba(99, 102, 241, 0.5);
    }

    .btn-primary:disabled {
      opacity: 0.6;
      cursor: not-allowed;
      transform: none;
    }

    .btn-secondary {
      background: rgba(255, 255, 255, 0.06);
      color: var(--text-primary);
      border: 1px solid var(--border-subtle);
    }

    .btn-secondary:hover {
      background: rgba(255, 255, 255, 0.1);
      border-color: var(--border-medium);
    }

    .btn-sm { padding: 5px 11px; font-size: 0.75rem; border-radius: var(--radius-sm); }

    /* Severity Distribution Visualizer */
    .distribution-panel {
      background: var(--bg-card);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg);
      padding: 16px 20px;
      margin-bottom: 24px;
    }

    .distribution-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 0.8rem;
      font-weight: 600;
      color: var(--text-secondary);
      margin-bottom: 10px;
    }

    .distribution-bar {
      height: 10px;
      border-radius: 9999px;
      background: rgba(255, 255, 255, 0.05);
      display: flex;
      overflow: hidden;
      gap: 2px;
    }

    .seg { height: 100%; transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
    .seg.crit { background: var(--sev-critical); }
    .seg.high { background: var(--sev-high); }
    .seg.med { background: var(--sev-medium); }
    .seg.low { background: var(--sev-low); }
    .seg.info { background: var(--sev-info); }

    /* Filter & Search Toolbar */
    .filter-section {
      display: flex;
      flex-direction: column;
      gap: 14px;
      margin-bottom: 20px;
    }

    .filter-row {
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      align-items: center;
      gap: 14px;
    }

    .pill-group {
      display: flex;
      flex-wrap: wrap;
      gap: 7px;
    }

    .filter-pill {
      padding: 6px 13px;
      border-radius: 9999px;
      font-size: 0.76rem;
      font-weight: 600;
      cursor: pointer;
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid var(--border-subtle);
      color: var(--text-secondary);
      transition: all 0.15s ease;
      display: inline-flex;
      align-items: center;
      gap: 5px;
    }

    .filter-pill:hover {
      color: var(--text-primary);
      border-color: rgba(255, 255, 255, 0.2);
    }

    .filter-pill.active {
      background: var(--accent-indigo);
      color: #ffffff;
      border-color: var(--accent-indigo);
      box-shadow: 0 2px 10px rgba(99, 102, 241, 0.3);
    }

    .search-container {
      flex: 1;
      min-width: 280px;
      max-width: 440px;
      position: relative;
    }

    .search-container input {
      width: 100%;
      padding: 8px 14px 8px 36px;
      border-radius: 9999px;
    }

    .search-icon {
      position: absolute;
      left: 12px;
      top: 50%;
      transform: translateY(-50%);
      color: var(--text-muted);
      pointer-events: none;
    }

    /* Findings List View */
    .findings-stream {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .finding-item {
      background: var(--bg-card);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-md);
      padding: 16px 20px;
      transition: all 0.2s ease;
      position: relative;
      border-left-width: 4px;
    }

    .finding-item:hover {
      background: var(--bg-card-hover);
      border-color: var(--border-medium);
      box-shadow: var(--shadow-sm);
    }

    .finding-item.critical { border-left-color: var(--sev-critical); }
    .finding-item.high { border-left-color: var(--sev-high); }
    .finding-item.medium { border-left-color: var(--sev-medium); }
    .finding-item.low { border-left-color: var(--sev-low); }
    .finding-item.info { border-left-color: var(--sev-info); }

    .item-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 14px;
      margin-bottom: 10px;
    }

    .item-tags {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
    }

    .badge-sev {
      font-size: 0.68rem;
      font-weight: 800;
      padding: 2px 8px;
      border-radius: 4px;
      text-transform: uppercase;
      letter-spacing: 0.06em;
    }

    .badge-sev.critical { background: var(--sev-critical-bg); color: var(--sev-critical); border: 1px solid var(--sev-critical-border); }
    .badge-sev.high { background: var(--sev-high-bg); color: var(--sev-high); border: 1px solid var(--sev-high-border); }
    .badge-sev.medium { background: var(--sev-medium-bg); color: var(--sev-medium); border: 1px solid var(--sev-medium-border); }
    .badge-sev.low { background: var(--sev-low-bg); color: var(--sev-low); border: 1px solid var(--sev-low-border); }
    .badge-sev.info { background: var(--sev-info-bg); color: var(--sev-info); border: 1px solid var(--sev-info-border); }

    .rule-title {
      font-size: 0.88rem;
      font-weight: 700;
      color: var(--text-primary);
    }

    .cwe-badge {
      font-size: 0.72rem;
      color: var(--text-muted);
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid var(--border-subtle);
      padding: 2px 7px;
      border-radius: 4px;
    }

    .engine-badge {
      font-size: 0.72rem;
      color: #a5b4fc;
      background: rgba(99, 102, 241, 0.1);
      padding: 2px 8px;
      border-radius: 4px;
    }

    .file-anchor {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-size: 0.8rem;
      color: #93c5fd;
      background: rgba(59, 130, 246, 0.08);
      border: 1px solid rgba(59, 130, 246, 0.2);
      padding: 3px 10px;
      border-radius: 6px;
      cursor: pointer;
      text-decoration: none;
      transition: background 0.15s ease;
    }

    .file-anchor:hover {
      background: rgba(59, 130, 246, 0.18);
    }

    .item-message {
      font-size: 0.9rem;
      color: #cbd5e1;
      margin-bottom: 12px;
      line-height: 1.45;
    }

    .item-actions {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-top: 12px;
      border-top: 1px solid rgba(255, 255, 255, 0.05);
    }

    .action-buttons {
      display: flex;
      gap: 8px;
    }

    /* Modal System */
    .modal-overlay {
      position: fixed;
      inset: 0;
      background: rgba(4, 7, 14, 0.82);
      backdrop-filter: blur(10px);
      -webkit-backdrop-filter: blur(10px);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1000;
      padding: 24px;
      opacity: 0;
      transition: opacity 0.2s ease;
    }

    .modal-overlay.active {
      display: flex;
      opacity: 1;
    }

    .modal-window {
      background: var(--bg-surface);
      border: 1px solid var(--border-medium);
      border-radius: var(--radius-xl);
      width: 100%;
      max-width: 900px;
      max-height: 88vh;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      box-shadow: var(--shadow-lg);
      transform: scale(0.96);
      transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    }

    .modal-overlay.active .modal-window {
      transform: scale(1);
    }

    .modal-top {
      padding: 20px 24px;
      border-bottom: 1px solid var(--border-subtle);
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: rgba(255, 255, 255, 0.02);
    }

    .modal-body {
      padding: 24px;
      overflow-y: auto;
      flex: 1;
    }

    .code-container {
      background: #050811;
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-sm);
      overflow-x: auto;
      padding: 12px 0;
      font-size: 0.84rem;
    }

    .code-row {
      display: flex;
      padding: 2px 16px;
    }

    .code-row.target {
      background: rgba(244, 63, 94, 0.18);
      border-left: 3px solid var(--sev-critical);
    }

    .gutter-num {
      width: 44px;
      color: var(--text-muted);
      text-align: right;
      margin-right: 18px;
      user-select: none;
    }

    .code-content {
      white-space: pre;
      color: #e2e8f0;
    }

    .rec-box {
      margin-top: 18px;
      padding: 16px;
      background: rgba(99, 102, 241, 0.06);
      border: 1px solid rgba(99, 102, 241, 0.2);
      border-radius: var(--radius-md);
    }

    .rec-box h4 {
      font-size: 0.84rem;
      color: #a5b4fc;
      margin-bottom: 6px;
    }

    .rec-box p {
      font-size: 0.82rem;
      color: #cbd5e1;
      line-height: 1.4;
    }

    /* Toast Notification */
    .toast-tray {
      position: fixed;
      bottom: 24px;
      right: 24px;
      display: flex;
      flex-direction: column;
      gap: 10px;
      z-index: 2000;
    }

    .toast {
      background: var(--bg-elevated);
      border: 1px solid var(--border-medium);
      color: var(--text-primary);
      padding: 12px 18px;
      border-radius: var(--radius-md);
      font-size: 0.82rem;
      font-weight: 500;
      box-shadow: var(--shadow-md);
      display: flex;
      align-items: center;
      gap: 10px;
      animation: slideIn 0.25s ease;
    }

    @keyframes slideIn {
      from { transform: translateX(40px); opacity: 0; }
      to { transform: translateX(0); opacity: 1; }
    }

    /* Empty States */
    .state-clean {
      text-align: center;
      padding: 64px 20px;
      background: var(--bg-card);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-lg);
    }

    .state-clean h3 {
      font-size: 1.3rem;
      font-weight: 700;
      color: var(--success-green);
      margin-bottom: 6px;
    }
  </style>
</head>
<body>
  <div class="app-container">
    <!-- Top Navigation -->
    <header>
      <div class="brand-section">
        <div class="brand-logo">🛡️</div>
        <div class="brand-titles">
          <h1>
            Go Code Scanner
            <span class="version-tag" id="toolVersion">v1.0.0</span>
          </h1>
          <p id="workspacePath">Workspace: Local Repository</p>
        </div>
      </div>
      <div class="header-controls">
        <div class="status-indicator">
          <span class="pulse-dot"></span>
          <span id="liveStatus">Engine Ready</span>
        </div>
        <button class="btn btn-secondary btn-sm" onclick="exportReportJSON()">💾 Export JSON</button>
      </div>
    </header>

    <!-- Executive Security Health & Metrics -->
    <div class="executive-grid">
      <!-- Security Gauge -->
      <div class="score-card">
        <div class="score-gauge">
          <svg viewBox="0 0 100 100">
            <circle class="bg-circle" cx="50" cy="50" r="38"></circle>
            <circle class="prog-circle" id="gaugeProgress" cx="50" cy="50" r="38"></circle>
          </svg>
          <div class="score-label">
            <div class="score-number" id="securityGrade">A</div>
            <div class="score-grade" id="securityScoreVal">100/100</div>
          </div>
        </div>
        <div class="score-details">
          <h3 id="healthTitle">Workspace Secure</h3>
          <p id="healthDesc">No active security policy violations detected.</p>
          <div class="policy-tag passed" id="policyStatusBadge">CI Policy: PASSING</div>
        </div>
      </div>

      <!-- KPI Grid -->
      <div class="kpi-grid">
        <div class="kpi-card total">
          <div class="kpi-title">Total Findings</div>
          <div class="kpi-value" id="kpiTotal">0</div>
          <div class="kpi-sub" id="kpiFilesMeta">0 files analyzed</div>
        </div>
        <div class="kpi-card critical">
          <div class="kpi-title">Critical</div>
          <div class="kpi-value" id="kpiCritical">0</div>
          <div class="kpi-sub">Immediate action required</div>
        </div>
        <div class="kpi-card high">
          <div class="kpi-title">High</div>
          <div class="kpi-value" id="kpiHigh">0</div>
          <div class="kpi-sub">High severity risks</div>
        </div>
        <div class="kpi-card medium">
          <div class="kpi-title">Medium</div>
          <div class="kpi-value" id="kpiMedium">0</div>
          <div class="kpi-sub">Potential vulnerabilities</div>
        </div>
        <div class="kpi-card low">
          <div class="kpi-title">Fixable ⚡</div>
          <div class="kpi-value" id="kpiFixable" style="color:#a5b4fc;">0</div>
          <div class="kpi-sub">1-Click auto-fixable</div>
        </div>
      </div>
    </div>

    <!-- Scan Control Bar -->
    <div class="scan-action-bar">
      <div class="form-item">
        <span>Mode:</span>
        <select id="scanMode">
          <option value="full">Full Working Tree</option>
          <option value="changed">Changed Files (HEAD)</option>
          <option value="staged">Staged Git Files</option>
        </select>
      </div>

      <div class="form-item">
        <span>Profile:</span>
        <select id="scanProfile">
          <option value="">Default Profile</option>
          <option value="strict">Strict Security</option>
          <option value="ci">CI Enforcement</option>
          <option value="fast">Fast Scan</option>
        </select>
      </div>

      <div class="form-item">
        <span>Scope:</span>
        <select id="scanScope">
          <option value="all">All Files</option>
          <option value="client">Client Only</option>
          <option value="server">Server Only</option>
        </select>
      </div>

      <label class="custom-checkbox">
        <input type="checkbox" id="autoFixCheckbox">
        <span>Auto-apply deterministic fixes</span>
      </label>

      <div style="margin-left: auto; display:flex; gap:10px;">
        <button class="btn btn-primary" id="btnExecuteScan" onclick="triggerScan()">
          <span id="scanSpinner" style="display:none;" class="spinner"></span>
          <span id="scanBtnLabel">⚡ Trigger Security Scan</span>
        </button>
      </div>
    </div>

    <!-- Severity Distribution Progress Bar -->
    <div class="distribution-panel">
      <div class="distribution-header">
        <span>Severity Distribution Breakdown</span>
        <span id="scanDurationText">Last scan: 0ms</span>
      </div>
      <div class="distribution-bar">
        <div class="seg crit" id="segCrit" style="width: 0%"></div>
        <div class="seg high" id="segHigh" style="width: 0%"></div>
        <div class="seg med" id="segMed" style="width: 0%"></div>
        <div class="seg low" id="segLow" style="width: 0%"></div>
        <div class="seg info" id="segInfo" style="width: 0%"></div>
      </div>
    </div>

    <!-- Filters & Search Toolbar -->
    <div class="filter-section">
      <div class="filter-row">
        <div class="pill-group" id="sevPills">
          <button class="filter-pill active" onclick="filterBySeverity('ALL')">All (<span id="cAll">0</span>)</button>
          <button class="filter-pill" onclick="filterBySeverity('CRITICAL')">Critical (<span id="cCrit">0</span>)</button>
          <button class="filter-pill" onclick="filterBySeverity('HIGH')">High (<span id="cHigh">0</span>)</button>
          <button class="filter-pill" onclick="filterBySeverity('MEDIUM')">Medium (<span id="cMed">0</span>)</button>
          <button class="filter-pill" onclick="filterBySeverity('LOW')">Low (<span id="cLow">0</span>)</button>
          <button class="filter-pill" onclick="filterBySeverity('FIXABLE')">Fixable ⚡ (<span id="cFix">0</span>)</button>
        </div>

        <div class="search-container">
          <svg class="search-icon" width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
          <input type="text" id="findingSearch" placeholder="Search by rule, file, CWE or keyword (⌘K)..." oninput="renderFindings()">
        </div>
      </div>
    </div>

    <!-- Findings Stream -->
    <div class="findings-stream" id="findingsStream">
      <div class="state-clean">
        <h3>✨ Ready to Analyze</h3>
        <p style="color:var(--text-secondary);font-size:0.88rem;margin-top:4px;">Klik "Trigger Security Scan" untuk menjalankan pemindaian pada workspace.</p>
      </div>
    </div>
  </div>

  <!-- Code Inspector Modal -->
  <div class="modal-overlay" id="codeModal">
    <div class="modal-window">
      <div class="modal-top">
        <div>
          <h2 style="font-size:1.1rem;font-weight:700;" id="codeModalRule">Rule Violation</h2>
          <p class="mono" id="codeModalPath" style="font-size:0.8rem;color:#93c5fd;margin-top:2px;"></p>
        </div>
        <button class="btn btn-secondary btn-sm" onclick="closeModal('codeModal')">✕ Close</button>
      </div>
      <div class="modal-body">
        <div class="code-container mono" id="codeBlock">
          <div style="padding:24px;text-align:center;color:var(--text-muted);">Fetching source lines...</div>
        </div>
        <div class="rec-box">
          <h4>💡 Remediation Advice</h4>
          <p id="codeModalAdvice"></p>
        </div>
      </div>
    </div>
  </div>

  <!-- Suppression Modal -->
  <div class="modal-overlay" id="suppressModal">
    <div class="modal-window" style="max-width:560px;">
      <div class="modal-top">
        <h2 style="font-size:1.1rem;font-weight:700;">🛡️ Add Suppression Rule</h2>
        <button class="btn btn-secondary btn-sm" onclick="closeModal('suppressModal')">✕</button>
      </div>
      <div class="modal-body">
        <form id="suppressForm" onsubmit="handleSuppressionSubmit(event)">
          <div style="margin-bottom:14px;">
            <label style="display:block;font-size:0.78rem;font-weight:600;color:var(--text-secondary);margin-bottom:5px;">Rule ID & Fingerprint</label>
            <input type="text" id="formRuleID" readonly style="width:100%;opacity:0.8;">
            <input type="hidden" id="formFingerprint">
          </div>
          <div style="margin-bottom:14px;">
            <label style="display:block;font-size:0.78rem;font-weight:600;color:var(--text-secondary);margin-bottom:5px;">File & Line Target</label>
            <input type="text" id="formFileLoc" readonly style="width:100%;opacity:0.8;">
          </div>
          <div style="margin-bottom:14px;">
            <label style="display:block;font-size:0.78rem;font-weight:600;color:var(--text-secondary);margin-bottom:5px;">Reason for Suppression (Required)</label>
            <textarea id="formReason" required rows="3" style="width:100%;background:var(--bg-input);border:1px solid var(--border-subtle);color:var(--text-primary);padding:8px 12px;border-radius:var(--radius-sm);font-family:inherit;font-size:0.85rem;" placeholder="State why this finding is an accepted risk or false positive..."></textarea>
          </div>
          <div style="margin-bottom:14px;">
            <label style="display:block;font-size:0.78rem;font-weight:600;color:var(--text-secondary);margin-bottom:5px;">Expiration Date (Required)</label>
            <input type="date" id="formExpires" required style="width:100%;">
          </div>
          <div style="margin-bottom:18px;">
            <label style="display:block;font-size:0.78rem;font-weight:600;color:var(--text-secondary);margin-bottom:5px;">Ticket / Reference (Optional)</label>
            <input type="text" id="formTicket" placeholder="e.g. SEC-892" style="width:100%;">
          </div>
          <div style="display:flex;justify-content:flex-end;gap:10px;">
            <button type="button" class="btn btn-secondary" onclick="closeModal('suppressModal')">Cancel</button>
            <button type="submit" class="btn btn-primary">Save Suppression</button>
          </div>
        </form>
      </div>
    </div>
  </div>

  <!-- Toast Notification Tray -->
  <div class="toast-tray" id="toastTray"></div>

  <script>
    let activeReport = null;
    let selectedSeverity = 'ALL';

    const defaultExp = new Date();
    defaultExp.setDate(defaultExp.getDate() + 90);
    document.getElementById('formExpires').value = defaultExp.toISOString().split('T')[0];

    window.addEventListener('DOMContentLoaded', () => {
      initApp();
      setupKeyboardShortcuts();
    });

    function setupKeyboardShortcuts() {
      window.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
          e.preventDefault();
          document.getElementById('findingSearch').focus();
        }
        if (e.key === 'Escape') {
          closeModal('codeModal');
          closeModal('suppressModal');
        }
      });
    }

    async function initApp() {
      try {
        const res = await fetch('/api/status');
        if (res.ok) {
          const status = await res.json();
          document.getElementById('workspacePath').innerText = 'Workspace: ' + status.root;
          if (status.tool_version) {
            document.getElementById('toolVersion').innerText = status.tool_version;
          }
        }
      } catch (err) {
        console.warn('Status error:', err);
      }
      loadReport();
    }

    async function loadReport() {
      try {
        const res = await fetch('/api/report');
        if (res.ok) {
          activeReport = await res.json();
          renderDashboard();
        }
      } catch (e) {
        console.log('No report yet.');
      }
    }

    async function triggerScan() {
      const btn = document.getElementById('btnExecuteScan');
      const label = document.getElementById('scanBtnLabel');
      const spinner = document.getElementById('scanSpinner');
      const live = document.getElementById('liveStatus');
      const mode = document.getElementById('scanMode').value;
      const profile = document.getElementById('scanProfile').value;
      const scope = document.getElementById('scanScope').value;
      const autoFix = document.getElementById('autoFixCheckbox').checked;

      btn.disabled = true;
      label.innerText = 'Scanning workspace...';
      live.innerText = 'Scanning...';
      live.style.color = '#f59e0b';

      try {
        const res = await fetch('/api/scan', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ mode, profile, scope, auto_fix: autoFix })
        });
        if (res.ok) {
          activeReport = await res.json();
          renderDashboard();
          live.innerText = 'Scan Synchronized';
          live.style.color = 'var(--success-green)';
          showToast('Security scan completed successfully!');
        } else {
          const err = await res.json();
          showToast('Scan error: ' + (err.error || 'Failed'), true);
        }
      } catch (e) {
        showToast('Request failed: ' + e.message, true);
      } finally {
        btn.disabled = false;
        label.innerText = '⚡ Trigger Security Scan';
      }
    }

    function renderDashboard() {
      if (!activeReport) return;
      const findings = activeReport.findings || [];

      let crit = 0, high = 0, med = 0, low = 0, info = 0, fixable = 0;
      findings.forEach(f => {
        const s = (f.severity || '').toUpperCase();
        if (s === 'CRITICAL') crit++;
        else if (s === 'HIGH') high++;
        else if (s === 'MEDIUM') med++;
        else if (s === 'LOW') low++;
        else info++;

        if (f.fixable || f.rule_id === 'trailing-whitespace' || f.rule_id === 'SQLI-001') fixable++;
      });

      document.getElementById('kpiTotal').innerText = findings.length;
      document.getElementById('kpiCritical').innerText = crit;
      document.getElementById('kpiHigh').innerText = high;
      document.getElementById('kpiMedium').innerText = med;
      document.getElementById('kpiFixable').innerText = fixable;

      document.getElementById('cAll').innerText = findings.length;
      document.getElementById('cCrit').innerText = crit;
      document.getElementById('cHigh').innerText = high;
      document.getElementById('cMed').innerText = med;
      document.getElementById('cLow').innerText = low;
      document.getElementById('cFix').innerText = fixable;

      // Security Health Score Calculation
      calculateSecurityScore(crit, high, med, low, findings.length);

      // Segment Progress Bar
      const total = findings.length || 1;
      document.getElementById('segCrit').style.width = (crit / total * 100) + '%';
      document.getElementById('segHigh').style.width = (high / total * 100) + '%';
      document.getElementById('segMed').style.width = (med / total * 100) + '%';
      document.getElementById('segLow').style.width = (low / total * 100) + '%';
      document.getElementById('segInfo').style.width = (info / total * 100) + '%';

      if (activeReport.metrics) {
        const dur = activeReport.metrics.duration_ms || 0;
        const files = activeReport.metrics.scanned_files || 0;
        document.getElementById('scanDurationText').innerText = 'Scanned ' + files + ' files in ' + dur + 'ms';
        document.getElementById('kpiFilesMeta').innerText = files + ' files scanned';
      }

      renderFindings();
    }

    function calculateSecurityScore(crit, high, med, low, total) {
      let score = 100;
      score -= (crit * 25);
      score -= (high * 10);
      score -= (med * 4);
      score -= (low * 1);
      if (score < 0) score = 0;

      let grade = 'A';
      let title = 'Workspace Secure';
      let desc = 'No critical or high security issues detected.';
      let color = 'var(--success-green)';
      let policyPass = true;

      if (crit > 0) {
        grade = 'F';
        title = 'Critical Vulnerabilities';
        desc = crit + ' critical issue(s) require immediate remediation.';
        color = 'var(--sev-critical)';
        policyPass = false;
      } else if (high > 0) {
        grade = 'C';
        title = 'High Risk Detected';
        desc = high + ' high severity vulnerability(ies) found.';
        color = 'var(--sev-high)';
        policyPass = false;
      } else if (med > 0) {
        grade = 'B';
        title = 'Moderate Risk';
        desc = med + ' medium severity findings to review.';
        color = 'var(--sev-medium)';
      }

      document.getElementById('securityGrade').innerText = grade;
      document.getElementById('securityGrade').style.color = color;
      document.getElementById('securityScoreVal').innerText = score + '/100';
      document.getElementById('healthTitle').innerText = title;
      document.getElementById('healthDesc').innerText = desc;

      const circle = document.getElementById('gaugeProgress');
      const offset = 238 - (238 * (score / 100));
      circle.style.strokeDashoffset = offset;
      circle.style.stroke = color;

      const policyBadge = document.getElementById('policyStatusBadge');
      if (policyPass) {
        policyBadge.className = 'policy-tag passed';
        policyBadge.innerText = 'CI Policy: PASSING';
      } else {
        policyBadge.className = 'policy-tag failed';
        policyBadge.innerText = 'CI Policy: BLOCKED';
      }
    }

    function filterBySeverity(sev) {
      selectedSeverity = sev;
      document.querySelectorAll('#sevPills .filter-pill').forEach(p => p.classList.remove('active'));
      if (event && event.currentTarget) {
        event.currentTarget.classList.add('active');
      }
      renderFindings();
    }

    function renderFindings() {
      const container = document.getElementById('findingsStream');
      if (!activeReport || !activeReport.findings || activeReport.findings.length === 0) {
        container.innerHTML = '<div class="state-clean"><h3>🎉 Clean Workspace</h3><p style="color:var(--text-secondary);font-size:0.88rem;">No security vulnerabilities found in current scan.</p></div>';
        return;
      }

      const q = (document.getElementById('findingSearch').value || '').toLowerCase();
      const filtered = activeReport.findings.filter(f => {
        const s = (f.severity || '').toUpperCase();
        if (selectedSeverity === 'CRITICAL' && s !== 'CRITICAL') return false;
        if (selectedSeverity === 'HIGH' && s !== 'HIGH') return false;
        if (selectedSeverity === 'MEDIUM' && s !== 'MEDIUM') return false;
        if (selectedSeverity === 'LOW' && s !== 'LOW') return false;
        if (selectedSeverity === 'FIXABLE' && !(f.fixable || f.rule_id === 'trailing-whitespace' || f.rule_id === 'SQLI-001')) return false;

        if (q) {
          const match = (f.rule_id || '').toLowerCase().includes(q) ||
                        (f.message || '').toLowerCase().includes(q) ||
                        (f.location && f.location.file && f.location.file.toLowerCase().includes(q)) ||
                        (f.cwe || '').toLowerCase().includes(q);
          if (!match) return false;
        }
        return true;
      });

      if (filtered.length === 0) {
        container.innerHTML = '<div class="state-clean"><h3>No Matching Findings</h3><p style="color:var(--text-secondary);font-size:0.88rem;">Try adjusting search keywords or severity filter.</p></div>';
        return;
      }

      let html = '';
      filtered.forEach((f) => {
        const sev = (f.severity || 'INFO').toUpperCase();
        const sevClass = sev.toLowerCase();
        const file = (f.location && f.location.file) ? f.location.file : 'Unknown';
        const line = (f.location && f.location.line) ? f.location.line : 1;
        const fileLoc = file + ':' + line;
        const isFixable = f.fixable || f.rule_id === 'trailing-whitespace' || f.rule_id === 'SQLI-001';
        const rawRule = f.rule_id || '';
        const rawRec = f.recommendation || f.message || '';
        const rawFp = f.fingerprint || '';

        html += '<div class="finding-item ' + sevClass + '">' +
          '<div class="item-header">' +
            '<div class="item-tags">' +
              '<span class="badge-sev ' + sevClass + '">' + escapeHtml(sev) + '</span>' +
              '<span class="rule-title mono">' + escapeHtml(rawRule) + '</span>' +
              (f.cwe ? '<span class="cwe-badge mono">' + escapeHtml(f.cwe) + '</span>' : '') +
              '<span class="engine-badge mono">' + escapeHtml(f.scanner || 'scanner') + '</span>' +
            '</div>' +
            '<a class="file-anchor mono" onclick="inspectCode(\'' + escapeAttr(file) + '\', ' + line + ', \'' + escapeAttr(rawRule) + '\', \'' + escapeAttr(rawRec) + '\')">' +
              '📄 ' + escapeHtml(fileLoc) +
            '</a>' +
          '</div>' +
          '<div class="item-message">' + escapeHtml(f.message) + '</div>' +
          '<div class="item-actions">' +
            '<div style="font-size:0.75rem;color:var(--text-muted);" class="mono">Fingerprint: ' + escapeHtml(rawFp.slice(0, 12)) + '...</div>' +
            '<div class="action-buttons">' +
              '<button class="btn btn-secondary btn-sm" onclick="inspectCode(\'' + escapeAttr(file) + '\', ' + line + ', \'' + escapeAttr(rawRule) + '\', \'' + escapeAttr(rawRec) + '\')">' +
                '🔍 Inspect Code' +
              '</button>' +
              (isFixable ? '<button class="btn btn-primary btn-sm" onclick="applySingleFix(\'' + escapeAttr(rawRule) + '\', \'' + escapeAttr(file) + '\', ' + line + ')">⚡ 1-Click Fix</button>' : '') +
              '<button class="btn btn-secondary btn-sm" onclick="openSuppressModal(\'' + escapeAttr(rawRule) + '\', \'' + escapeAttr(rawFp) + '\', \'' + escapeAttr(file) + '\', ' + line + ')">' +
                '🛡️ Suppress' +
              '</button>' +
            '</div>' +
          '</div>' +
        '</div>';
      });
      container.innerHTML = html;
    }

    async function inspectCode(filePath, targetLine, ruleId, recommendation) {
      const modal = document.getElementById('codeModal');
      document.getElementById('codeModalRule').innerText = 'Rule Violation: ' + ruleId;
      document.getElementById('codeModalPath').innerText = filePath + ' (Line ' + targetLine + ')';
      document.getElementById('codeModalAdvice').innerText = recommendation || 'Review code context and apply appropriate security hardening.';

      const block = document.getElementById('codeBlock');
      block.innerHTML = '<div style="padding:24px;text-align:center;color:var(--text-muted);">Reading file lines...</div>';
      modal.classList.add('active');

      try {
        const res = await fetch('/api/file?path=' + encodeURIComponent(filePath));
        if (!res.ok) throw new Error('Could not access file on disk');
        const data = await res.json();

        const start = Math.max(1, targetLine - 7);
        const end = Math.min(data.total_lines, targetLine + 7);
        let rows = '';

        for (let i = start; i <= end; i++) {
          const text = data.lines[i - 1] || '';
          const isTarget = (i === targetLine);
          rows += '<div class="code-row ' + (isTarget ? 'target' : '') + '">' +
            '<div class="gutter-num">' + i + '</div>' +
            '<div class="code-content">' + escapeHtml(text) + '</div>' +
          '</div>';
        }
        block.innerHTML = rows;
      } catch (err) {
        block.innerHTML = '<div style="padding:20px;color:var(--sev-critical);">Failed to load source code: ' + escapeHtml(err.message) + '</div>';
      }
    }

    function openSuppressModal(ruleId, fingerprint, file, line) {
      document.getElementById('formRuleID').value = ruleId;
      document.getElementById('formFingerprint').value = fingerprint;
      document.getElementById('formFileLoc').value = file + ':' + line;
      document.getElementById('formReason').value = '';
      document.getElementById('suppressModal').classList.add('active');
    }

    async function handleSuppressionSubmit(e) {
      e.preventDefault();
      const rule_id = document.getElementById('formRuleID').value;
      const fingerprint = document.getElementById('formFingerprint').value;
      const locParts = document.getElementById('formFileLoc').value.split(':');
      const file = locParts[0];
      const line = parseInt(locParts[1] || '0', 10);
      const reason = document.getElementById('formReason').value;
      const expires = document.getElementById('formExpires').value;
      const ticket = document.getElementById('formTicket').value;

      try {
        const res = await fetch('/api/suppress', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ rule_id, fingerprint, file, line, reason, expires, ticket })
        });
        if (res.ok) {
          const data = await res.json();
          activeReport = data.report;
          closeModal('suppressModal');
          renderDashboard();
          showToast('Suppression rule applied! Workspace rescanned.');
        } else {
          const err = await res.json();
          showToast('Suppression failed: ' + (err.error || 'Unknown error'), true);
        }
      } catch (err) {
        showToast('Request error: ' + err.message, true);
      }
    }

    async function applySingleFix(rule_id, file, line) {
      try {
        const res = await fetch('/api/fix', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ rule_id, file, line })
        });
        if (res.ok) {
          const data = await res.json();
          activeReport = data.report;
          renderDashboard();
          showToast('Deterministic fix applied to ' + file + ':' + line);
        } else {
          const err = await res.json();
          showToast('Fix failed: ' + (err.error || 'Unknown error'), true);
        }
      } catch (err) {
        showToast('Fix error: ' + err.message, true);
      }
    }

    function closeModal(id) {
      document.getElementById(id).classList.remove('active');
    }

    function showToast(msg, isError = false) {
      const tray = document.getElementById('toastTray');
      const toast = document.createElement('div');
      toast.className = 'toast';
      if (isError) {
        toast.style.borderColor = 'var(--sev-critical-border)';
        toast.innerHTML = '⚠️ <span>' + escapeHtml(msg) + '</span>';
      } else {
        toast.style.borderColor = 'rgba(16,185,129,0.4)';
        toast.innerHTML = '✅ <span>' + escapeHtml(msg) + '</span>';
      }
      tray.appendChild(toast);
      setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(10px)';
        toast.style.transition = 'all 0.3s ease';
        setTimeout(() => toast.remove(), 300);
      }, 4000);
    }

    function exportReportJSON() {
      if (!activeReport) return showToast('No scan report to export', true);
      const blob = new Blob([JSON.stringify(activeReport, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'security-report-' + new Date().toISOString().replace(/[:.]/g, '-') + '.json';
      a.click();
      URL.revokeObjectURL(url);
      showToast('Exported report to JSON');
    }

    function escapeHtml(str) {
      if (!str) return '';
      return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
    }

    function escapeAttr(str) {
      if (!str) return '';
      return String(str)
        .replace(/\\/g, '\\\\')
        .replace(/'/g, "\\'")
        .replace(/"/g, '&quot;');
    }
  </script>
</body>
</html>`
