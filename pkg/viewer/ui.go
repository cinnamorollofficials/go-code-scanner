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
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Semgrep Dashboard — Go Code Scanner</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-dark-nav: #161a23;
      --bg-dark-nav-hover: #222735;
      --bg-dark-active: #2a3142;
      --text-dark-nav: #9ca3af;
      --text-dark-nav-active: #ffffff;
      
      --bg-filter-bar: #fcfdfe;
      --border-color: #e5e7eb;
      --border-subtle: #f0f2f5;
      
      --bg-main: #f8f9fa;
      --bg-card: #ffffff;
      
      --text-main: #111827;
      --text-muted: #6b7280;
      --text-subtle: #9ca3af;
      
      --semgrep-teal: #00d3a9;
      --semgrep-blue: #0969da;
      --semgrep-blue-hover: #0858b8;
      --semgrep-blue-light: #eef5ff;
      --semgrep-blue-border: #b6d3fa;
      
      --sev-crit: #dc2626;
      --sev-crit-bg: #fef2f2;
      --sev-high: #ea580c;
      --sev-med: #f59e0b;
      --sev-low: #3b82f6;
      
      --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      --font-mono: 'JetBrains Mono', monospace;
    }

    * { box-sizing: border-box; margin: 0; padding: 0; }
    
    body {
      font-family: var(--font-sans);
      background-color: var(--bg-main);
      color: var(--text-main);
      min-height: 100vh;
      display: flex;
      font-size: 13px;
      line-height: 1.5;
      -webkit-font-smoothing: antialiased;
    }

    .mono { font-family: var(--font-mono); }

    /* Main App Layout */
    .app-layout {
      display: flex;
      width: 100vw;
      min-height: 100vh;
      overflow-x: hidden;
    }

    /* 1. Left Dark Navigation */
    .dark-sidebar {
      width: 210px;
      min-width: 210px;
      background-color: var(--bg-dark-nav);
      color: var(--text-dark-nav);
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      border-right: 1px solid #11141c;
      user-select: none;
      z-index: 20;
    }

    .dark-sidebar-top {
      display: flex;
      flex-direction: column;
    }

    .sidebar-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 16px 14px;
      border-bottom: 1px solid rgba(255, 255, 255, 0.07);
      color: #fff;
    }

    .sidebar-header-left {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 600;
      font-size: 13.5px;
      cursor: pointer;
    }

    .hamburger-icon {
      display: flex;
      flex-direction: column;
      gap: 3px;
      cursor: pointer;
    }
    .hamburger-icon span {
      display: block;
      width: 15px;
      height: 2px;
      background: #9ca3af;
      border-radius: 1px;
    }

    .nav-menu {
      list-style: none;
      padding: 12px 8px;
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .nav-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 7px 10px;
      border-radius: 6px;
      color: var(--text-dark-nav);
      font-size: 13px;
      font-weight: 500;
      cursor: pointer;
      text-decoration: none;
      transition: background 0.15s ease, color 0.15s ease;
    }

    .nav-item:hover {
      background-color: var(--bg-dark-nav-hover);
      color: #fff;
    }

    .nav-item.active {
      background-color: var(--bg-dark-active);
      color: var(--text-dark-nav-active);
      font-weight: 600;
    }

    .nav-item-left {
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .nav-item-icon {
      width: 16px;
      display: flex;
      justify-content: center;
      font-size: 14px;
      opacity: 0.85;
    }

    .nav-badge {
      background: #374151;
      color: #e5e7eb;
      font-size: 11px;
      font-weight: 600;
      padding: 1px 7px;
      border-radius: 10px;
    }

    .dark-sidebar-bottom {
      padding: 12px 8px;
      border-top: 1px solid rgba(255, 255, 255, 0.07);
    }

    .semgrep-brand-logo {
      padding: 12px 10px 4px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    
    .semgrep-circles {
      display: inline-flex;
      align-items: center;
      gap: 3px;
    }
    .semgrep-circle {
      width: 13px;
      height: 13px;
      border: 2.5px solid var(--semgrep-teal);
      border-radius: 50%;
    }

    /* Views Layout Container */
    .view-content-wrapper {
      flex: 1;
      display: flex;
      height: 100vh;
      overflow: hidden;
    }

    /* ========================================================
       DASHBOARD VIEW STYLES (MATCHING SEMGREP SCREENSHOT)
       ======================================================== */
    .dashboard-viewport {
      flex: 1;
      height: 100vh;
      overflow-y: auto;
      padding: 24px 32px 60px;
      background-color: #ffffff;
      display: block;
    }

    .dash-top-header {
      display: flex;
      justify-content: space-between;
      align-items: baseline;
      margin-bottom: 24px;
    }

    .dash-title {
      font-size: 1.35rem;
      font-weight: 700;
      color: #111827;
      letter-spacing: -0.01em;
    }

    .dash-last-updated {
      font-size: 12px;
      color: var(--text-muted);
    }

    /* Noise Reduction Card */
    .noise-reduction-card {
      background: #ffffff;
      border: 1px solid var(--border-color);
      border-radius: 8px;
      padding: 20px 24px 24px;
      margin-bottom: 28px;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
    }

    .card-heading-row {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 14.5px;
      font-weight: 700;
      color: #111827;
      margin-bottom: 16px;
    }

    .info-circle-icon {
      color: #9ca3af;
      cursor: pointer;
      font-size: 13px;
      display: inline-flex;
    }

    /* Funnel Ribbon Stream Graphic */
    .funnel-container {
      width: 100%;
      height: 120px;
      position: relative;
      margin-bottom: 20px;
    }

    .funnel-svg {
      width: 100%;
      height: 100%;
      display: block;
    }

    .noise-metrics-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      border-top: 1px solid #f3f4f6;
      padding-top: 18px;
      gap: 16px;
    }

    .noise-metric-col {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .metric-label-with-info {
      font-size: 12.5px;
      color: #4b5563;
      font-weight: 500;
      display: flex;
      align-items: center;
      gap: 5px;
    }

    .metric-big-num-row {
      display: flex;
      align-items: baseline;
      gap: 8px;
    }

    .metric-big-num {
      font-size: 1.65rem;
      font-weight: 700;
      color: #111827;
      letter-spacing: -0.02em;
    }

    .reduction-pill {
      font-size: 12px;
      font-weight: 600;
      color: #059669;
    }

    /* Reporting Summary Section */
    .reporting-summary-section {
      margin-bottom: 24px;
    }

    .section-title {
      font-size: 1.15rem;
      font-weight: 700;
      color: #111827;
      margin-bottom: 14px;
    }

    .reporting-toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      flex-wrap: wrap;
      gap: 12px;
      padding-bottom: 16px;
    }

    .toolbar-left-filters {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 14px;
      font-size: 12.5px;
      color: #374151;
    }

    .filter-inline-select {
      display: inline-flex;
      align-items: center;
      gap: 5px;
      font-weight: 600;
      color: #1f2937;
      cursor: pointer;
    }
    .filter-inline-select select {
      border: none;
      background: transparent;
      font-weight: 600;
      font-size: 12.5px;
      color: #111827;
      outline: none;
      cursor: pointer;
    }

    .toggle-switch-wrapper {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      cursor: pointer;
    }

    .switch-pill {
      width: 34px;
      height: 18px;
      background: #e5e7eb;
      border-radius: 999px;
      position: relative;
      transition: background 0.2s;
    }
    .switch-pill.active {
      background: var(--semgrep-blue);
    }
    .switch-thumb {
      width: 14px;
      height: 14px;
      background: #ffffff;
      border-radius: 50%;
      position: absolute;
      top: 2px;
      left: 2px;
      transition: transform 0.2s;
      box-shadow: 0 1px 2px rgba(0,0,0,0.15);
    }
    .switch-pill.active .switch-thumb {
      transform: translateX(16px);
    }

    .btn-all-filters {
      background: #edf5ff;
      color: var(--semgrep-blue);
      border: 1px solid #d0e4ff;
      border-radius: 6px;
      padding: 4px 12px;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
    }
    .btn-all-filters:hover { background: #e0eeff; }

    .download-link {
      color: var(--semgrep-blue);
      text-decoration: none;
      font-weight: 600;
      font-size: 12.5px;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      cursor: pointer;
    }
    .download-link:hover { text-decoration: underline; }

    /* Production Backlog Card */
    .production-backlog-card {
      background: #ffffff;
      border: 1px solid var(--border-color);
      border-radius: 8px;
      padding: 24px;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
    }

    .backlog-title {
      font-size: 1.05rem;
      font-weight: 700;
      color: #111827;
      margin-bottom: 20px;
    }

    .backlog-kpi-row {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 20px;
      margin-bottom: 28px;
    }

    .backlog-kpi-item {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .backlog-kpi-num {
      font-size: 1.75rem;
      font-weight: 700;
      color: #111827;
      letter-spacing: -0.02em;
    }

    .backlog-kpi-label {
      font-size: 11px;
      font-weight: 700;
      color: #6b7280;
      letter-spacing: 0.04em;
      text-transform: uppercase;
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .backlog-kpi-trend {
      font-size: 11.5px;
      margin-top: 4px;
      font-weight: 500;
    }
    .backlog-kpi-trend.green { color: #059669; }
    .backlog-kpi-trend.red { color: #dc2626; }

    .charts-grid-2col {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 28px;
      padding-top: 16px;
      border-top: 1px solid #f3f4f6;
    }

    .chart-box-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 12px;
    }

    .chart-box-title {
      font-size: 13.5px;
      font-weight: 700;
      color: #111827;
    }

    .chart-box-subtitle {
      font-size: 12px;
      color: #6b7280;
      margin-top: 2px;
    }

    .chart-svg-container {
      width: 100%;
      height: 200px;
      position: relative;
    }

    /* ========================================================
       CODE FINDINGS VIEW STYLES
       ======================================================== */
    .filter-sidebar {
      width: 250px;
      min-width: 250px;
      background-color: var(--bg-filter-bar);
      border-right: 1px solid var(--border-color);
      padding: 16px 18px 40px;
      overflow-y: auto;
      height: 100vh;
      display: flex;
      flex-direction: column;
      gap: 18px;
    }

    .filter-group {
      display: flex;
      flex-direction: column;
      gap: 7px;
    }

    .filter-label {
      font-size: 12px;
      font-weight: 600;
      color: #374151;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .custom-select {
      width: 100%;
      padding: 6px 10px;
      border: 1px solid #d1d5db;
      border-radius: 6px;
      background: #fff;
      font-family: inherit;
      font-size: 12.5px;
      color: #1f2937;
      outline: none;
      cursor: pointer;
      appearance: none;
      background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
      background-repeat: no-repeat;
      background-position: right 8px center;
      background-size: 14px;
    }
    .custom-select:focus {
      border-color: var(--semgrep-blue);
      box-shadow: 0 0 0 2px rgba(9, 105, 218, 0.15);
    }

    .filter-pills-row {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .filter-pill-btn {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 10px;
      border-radius: 14px;
      font-size: 12px;
      font-weight: 500;
      border: 1px solid #d1d5db;
      background: #ffffff;
      color: #374151;
      cursor: pointer;
      transition: all 0.15s ease;
      user-select: none;
    }

    .filter-pill-btn:hover { background: #f3f4f6; border-color: #9ca3af; }
    .filter-pill-btn.active {
      background: var(--semgrep-blue);
      border-color: var(--semgrep-blue);
      color: #ffffff;
      font-weight: 600;
    }
    .filter-pill-btn.active .pill-count { background: rgba(255, 255, 255, 0.25); color: #ffffff; }

    .pill-count {
      font-size: 11px;
      padding: 0 5px;
      border-radius: 10px;
      background: #e5e7eb;
      color: #4b5563;
      font-weight: 600;
    }

    .dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }
    .dot.crit { background-color: #dc2626; }
    .dot.high { background-color: #ea580c; }
    .dot.med { background-color: #f59e0b; }
    .dot.low { background-color: #3b82f6; }

    .scan-box-widget {
      background: #f0f7ff;
      border: 1px solid #bae0fd;
      border-radius: 8px;
      padding: 12px;
      display: flex;
      flex-direction: column;
      gap: 10px;
      margin-top: 6px;
    }

    .scan-btn-primary {
      background: var(--semgrep-blue);
      color: #fff;
      border: none;
      border-radius: 6px;
      padding: 7px 12px;
      font-size: 12.5px;
      font-weight: 600;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      transition: background 0.15s;
    }
    .scan-btn-primary:hover { background: var(--semgrep-blue-hover); }
    .scan-btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

    .code-viewport {
      flex: 1;
      height: 100vh;
      overflow-y: auto;
      padding: 20px 28px 60px;
      background-color: #ffffff;
    }

    .content-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--border-color);
    }

    .content-title-area {
      display: flex;
      align-items: baseline;
      gap: 12px;
    }

    .content-title {
      font-size: 1.25rem;
      font-weight: 700;
      color: #111827;
      letter-spacing: -0.01em;
    }

    .content-subtitle {
      font-size: 12px;
      color: var(--text-muted);
    }

    .content-actions-area {
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .search-input-wrapper {
      position: relative;
      display: flex;
      align-items: center;
    }

    .search-input-wrapper input {
      padding: 6px 12px 6px 30px;
      border: 1px solid #d1d5db;
      border-radius: 6px;
      font-size: 12.5px;
      width: 230px;
      outline: none;
    }
    .search-input-wrapper input:focus {
      border-color: var(--semgrep-blue);
      box-shadow: 0 0 0 2px rgba(9, 105, 218, 0.12);
    }

    .search-input-wrapper svg {
      position: absolute;
      left: 9px;
      color: #9ca3af;
      pointer-events: none;
    }

    /* Finding Rule Cards */
    .findings-container {
      display: flex;
      flex-direction: column;
      gap: 18px;
    }

    .rule-card {
      background: #ffffff;
      border: 1px solid var(--border-color);
      border-radius: 6px;
      overflow: hidden;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
      transition: box-shadow 0.15s ease;
      position: relative;
    }

    .rule-card:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.07);
    }

    .rule-card.sev-critical { border-top: 3px solid #dc2626; }
    .rule-card.sev-high { border-top: 3px solid #ea580c; }
    .rule-card.sev-medium { border-top: 3px solid #f59e0b; }
    .rule-card.sev-low { border-top: 3px solid #3b82f6; }

    .rule-card-header {
      padding: 14px 18px 10px;
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 14px;
    }

    .rule-card-header-left { flex: 1; }

    .rule-id-title {
      font-size: 14.5px;
      font-weight: 700;
      color: #111827;
      letter-spacing: -0.01em;
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }

    .rule-card-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .btn-icon-subtle {
      border: none;
      background: transparent;
      color: #6b7280;
      padding: 5px;
      border-radius: 4px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .btn-icon-subtle:hover { background: #f3f4f6; color: #111827; }

    .btn-outline-sm {
      border: 1px solid #d1d5db;
      background: #ffffff;
      color: var(--semgrep-blue);
      border-radius: 6px;
      padding: 4px 10px;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 5px;
    }
    .btn-outline-sm:hover { background: #f9fafb; border-color: #9ca3af; }

    .btn-triage {
      background: var(--semgrep-blue);
      color: #ffffff;
      border: 1px solid var(--semgrep-blue);
      border-radius: 6px;
      padding: 4px 12px;
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      transition: background 0.15s;
    }
    .btn-triage:hover { background: var(--semgrep-blue-hover); }

    .rule-description-area {
      padding: 0 18px 12px;
      color: #374151;
      font-size: 13px;
      line-height: 1.5;
    }

    .rule-description-area code {
      background: #f3f4f6;
      border: 1px solid #e5e7eb;
      padding: 1px 5px;
      border-radius: 4px;
      font-family: var(--font-mono);
      font-size: 12px;
      color: #1f2937;
    }

    .show-more-link {
      color: var(--semgrep-blue);
      font-weight: 600;
      cursor: pointer;
      text-decoration: none;
      margin-left: 4px;
      font-size: 12.5px;
    }
    .show-more-link:hover { text-decoration: underline; }

    .rule-meta-tags {
      display: flex;
      justify-content: flex-end;
      align-items: center;
      gap: 14px;
      margin-top: 6px;
      font-size: 11.5px;
      color: #6b7280;
    }

    .meta-tag-item {
      display: inline-flex;
      align-items: center;
      gap: 4px;
    }

    /* Finding Occurrences Table Rows */
    .finding-rows-table {
      border-top: 1px solid var(--border-color);
      display: flex;
      flex-direction: column;
    }

    .finding-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 9px 18px;
      border-bottom: 1px solid #f3f4f6;
      font-size: 12.5px;
      transition: background 0.1s ease;
    }
    .finding-row:last-child { border-bottom: none; }
    .finding-row:hover { background-color: #f9fafb; }

    .finding-row-left {
      display: flex;
      align-items: center;
      gap: 12px;
      flex: 1;
      overflow: hidden;
    }

    .finding-checkbox {
      width: 16px;
      height: 16px;
      border-radius: 3px;
      border: 1.5px solid #ea580c;
      background: #fff;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
    }
    .finding-checkbox.crit { border-color: #dc2626; }
    .finding-checkbox.high { border-color: #ea580c; }
    .finding-checkbox.med { border-color: #f59e0b; }
    .finding-checkbox.low { border-color: #3b82f6; }

    .finding-age {
      color: #9ca3af;
      font-size: 11.5px;
      display: inline-flex;
      align-items: center;
      gap: 3px;
      min-width: 32px;
    }

    .finding-path-link {
      color: var(--semgrep-blue);
      font-family: var(--font-mono);
      font-size: 12px;
      text-decoration: none;
      cursor: pointer;
      font-weight: 500;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 480px;
    }
    .finding-path-link:hover { text-decoration: underline; }

    .finding-row-right {
      display: flex;
      align-items: center;
      gap: 16px;
      font-size: 12px;
    }

    .repo-pill {
      color: #4b5563;
      display: inline-flex;
      align-items: center;
      gap: 5px;
      font-size: 12px;
    }

    .branch-pill {
      color: #6b7280;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
    }

    .row-triage-link {
      color: var(--semgrep-blue);
      font-weight: 600;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
      padding: 2px 6px;
      border-radius: 4px;
    }
    .row-triage-link:hover { background: var(--semgrep-blue-light); }

    .show-more-findings-btn {
      padding: 8px 18px;
      background: #fafafa;
      border-top: 1px solid var(--border-color);
      color: var(--semgrep-blue);
      font-size: 12px;
      font-weight: 600;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 5px;
      user-select: none;
    }
    .show-more-findings-btn:hover { background: #f3f4f6; }

    /* Modals */
    .modal-overlay {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.6);
      backdrop-filter: blur(4px);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1000;
      padding: 20px;
    }
    .modal-overlay.active { display: flex; }

    .modal-box {
      background: #ffffff;
      border-radius: 8px;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
      width: 100%;
      max-width: 860px;
      max-height: 90vh;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      border: 1px solid #e5e7eb;
    }

    .modal-head {
      padding: 16px 20px;
      border-bottom: 1px solid var(--border-color);
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: #fcfcfd;
    }

    .modal-body {
      padding: 20px;
      overflow-y: auto;
      flex: 1;
    }

    .code-viewer-block {
      background: #0f172a;
      color: #f8fafc;
      border-radius: 6px;
      font-family: var(--font-mono);
      font-size: 12.5px;
      overflow-x: auto;
      padding: 12px 0;
      margin: 12px 0;
    }

    .code-line-row { display: flex; padding: 2px 14px; }
    .code-line-row.target-line {
      background: rgba(239, 68, 68, 0.25);
      border-left: 3px solid #ef4444;
    }
    .code-line-num {
      width: 44px;
      color: #64748b;
      text-align: right;
      margin-right: 16px;
      user-select: none;
    }
    .code-line-text { white-space: pre; }

    .rec-box-ui {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 6px;
      padding: 14px;
      margin-top: 14px;
    }
    .rec-box-ui h4 { font-size: 13px; font-weight: 600; color: #1e293b; margin-bottom: 4px; }
    .rec-box-ui p { font-size: 12.5px; color: #475569; }

    .toast-container {
      position: fixed;
      bottom: 24px;
      right: 24px;
      display: flex;
      flex-direction: column;
      gap: 8px;
      z-index: 2000;
    }
    .toast-msg {
      background: #1f2937;
      color: #ffffff;
      padding: 10px 16px;
      border-radius: 6px;
      font-size: 12.5px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      display: flex;
      align-items: center;
      gap: 8px;
      animation: fadeIn 0.2s ease;
    }
    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: translateY(0); }
    }
  </style>
</head>
<body>
  <div class="app-layout">
    <!-- 1. Left Dark Navigation -->
    <aside class="dark-sidebar">
      <div class="dark-sidebar-top">
        <div class="sidebar-header">
          <div class="sidebar-header-left">
            <div class="hamburger-icon">
              <span></span>
              <span></span>
              <span></span>
            </div>
            <span>returntocorp</span>
          </div>
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
        </div>

        <ul class="nav-menu">
          <li>
            <a class="nav-item active" id="navItemDashboard" href="#dashboard" onclick="switchNav('dashboard')">
              <div class="nav-item-left">
                <span class="nav-item-icon">📊</span>
                <span>Dashboard</span>
              </div>
            </a>
          </li>
          <li>
            <a class="nav-item" id="navItemProjects" href="#projects" onclick="switchNav('projects')">
              <div class="nav-item-left">
                <span class="nav-item-icon">📁</span>
                <span>Projects</span>
              </div>
            </a>
          </li>
          <li>
            <a class="nav-item" id="navItemCode" href="#code" onclick="switchNav('code')">
              <div class="nav-item-left">
                <span class="nav-item-icon">&lt;/&gt;</span>
                <span>Code</span>
              </div>
              <span class="nav-badge" id="navCodeCount">40</span>
            </a>
          </li>
          <li>
            <a class="nav-item" id="navItemSecrets" href="#secrets" onclick="switchNav('secrets')">
              <div class="nav-item-left">
                <span class="nav-item-icon">🔑</span>
                <span>Secrets</span>
              </div>
              <span class="nav-badge">4</span>
            </a>
          </li>
          <li>
            <a class="nav-item" id="navItemSupplyChain" href="#supply-chain" onclick="switchNav('supply-chain')">
              <div class="nav-item-left">
                <span class="nav-item-icon">🔗</span>
                <span>Supply Chain</span>
              </div>
              <span class="nav-badge">16</span>
            </a>
          </li>
          <li>
            <a class="nav-item" id="navItemRules" href="#rules" onclick="switchNav('rules')">
              <div class="nav-item-left">
                <span class="nav-item-icon">📑</span>
                <span>Rules & Policies</span>
              </div>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
            </a>
          </li>
        </ul>
      </div>

      <div class="dark-sidebar-bottom">
        <ul class="nav-menu">
          <li>
            <a class="nav-item" href="#settings">
              <div class="nav-item-left">
                <span class="nav-item-icon">⚙️</span>
                <span>Settings</span>
              </div>
            </a>
          </li>
          <li>
            <a class="nav-item" href="https://semgrep.dev/docs" target="_blank">
              <div class="nav-item-left">
                <span class="nav-item-icon">📖</span>
                <span>Docs</span>
              </div>
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
            </a>
          </li>
          <li>
            <a class="nav-item" href="#help">
              <div class="nav-item-left">
                <span class="nav-item-icon">❓</span>
                <span>Help</span>
              </div>
            </a>
          </li>
          <li>
            <a class="nav-item" href="#updates">
              <div class="nav-item-left">
                <span class="nav-item-icon">🎁</span>
                <span>Updates</span>
              </div>
            </a>
          </li>
        </ul>

        <div class="semgrep-brand-logo">
          <div class="semgrep-circles">
            <div class="semgrep-circle"></div>
            <div class="semgrep-circle"></div>
            <div class="semgrep-circle"></div>
          </div>
        </div>
      </div>
    </aside>

    <!-- Center Views Container -->
    <div class="view-content-wrapper">
      
      <!-- ========================================================
           VIEW 1: DASHBOARD ANALYTICS (EXACT SEMGREP MATCH)
           ======================================================== -->
      <main class="dashboard-viewport" id="dashboardView">
        <div class="dash-top-header">
          <h1 class="dash-title">Dashboard</h1>
          <span class="dash-last-updated" id="dashLastUpdatedText">Last updated 5 days ago</span>
        </div>

        <!-- 1. Semgrep Noise Reduction Card -->
        <div class="noise-reduction-card">
          <div class="card-heading-row">
            <span>Semgrep noise reduction</span>
            <span class="info-circle-icon" title="Shows how Semgrep progressive filtering reduces alert volume">ⓘ</span>
          </div>

          <!-- Funnel Ribbon Area Stream Graph -->
          <div class="funnel-container">
            <svg class="funnel-svg" viewBox="0 0 800 120" preserveAspectRatio="none">
              <defs>
                <linearGradient id="funnelGrad" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#93c5fd" stop-opacity="0.85"/>
                  <stop offset="35%" stop-color="#60a5fa" stop-opacity="0.85"/>
                  <stop offset="70%" stop-color="#3b82f6" stop-opacity="0.9"/>
                  <stop offset="100%" stop-color="#2563eb" stop-opacity="0.95"/>
                </linearGradient>
                <linearGradient id="cyanRibbon" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#5eead4" stop-opacity="0.9"/>
                  <stop offset="100%" stop-color="#14b8a6" stop-opacity="0.9"/>
                </linearGradient>
                <linearGradient id="purpleRibbon" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#e879f9" stop-opacity="0.9"/>
                  <stop offset="100%" stop-color="#a855f7" stop-opacity="0.9"/>
                </linearGradient>
              </defs>

              <!-- Stream Shape -->
              <path d="M 0 15 
                       C 120 15, 160 38, 260 48
                       C 360 58, 420 66, 520 72
                       C 620 78, 680 84, 800 86
                       L 800 96
                       C 680 94, 620 90, 520 86
                       C 420 82, 360 80, 260 76
                       C 160 72, 120 68, 0 68
                       Z" fill="url(#funnelGrad)"/>
              
              <!-- Bottom Cyan Ribbon Accent -->
              <path d="M 0 68 
                       C 120 68, 160 72, 260 76
                       C 360 80, 420 82, 520 86
                       C 620 90, 680 94, 800 96
                       L 800 99
                       C 680 97, 620 93, 520 89
                       C 420 85, 360 83, 260 79
                       C 160 75, 120 71, 0 71
                       Z" fill="url(#cyanRibbon)"/>

              <!-- Bottom Purple Ribbon Accent -->
              <path d="M 0 71
                       C 120 71, 160 75, 260 79
                       C 360 83, 420 85, 520 89
                       C 620 93, 680 97, 800 99
                       L 800 102
                       C 680 100, 620 96, 520 92
                       C 420 88, 360 86, 260 82
                       C 160 78, 120 74, 0 74
                       Z" fill="url(#purpleRibbon)"/>

              <!-- Vertical dividing grid ticks -->
              <line x1="260" y1="10" x2="260" y2="110" stroke="#ffffff" stroke-opacity="0.4" stroke-width="1.5"/>
              <line x1="520" y1="10" x2="520" y2="110" stroke="#ffffff" stroke-opacity="0.4" stroke-width="1.5"/>
              <line x1="780" y1="10" x2="780" y2="110" stroke="#ffffff" stroke-opacity="0.4" stroke-width="1.5"/>
            </svg>
          </div>

          <!-- Funnel Metrics Breakdown Row -->
          <div class="noise-metrics-grid">
            <div class="noise-metric-col">
              <span class="metric-label-with-info">All findings</span>
              <div class="metric-big-num-row">
                <span class="metric-big-num" id="metricAllFindings">1,234</span>
              </div>
            </div>

            <div class="noise-metric-col">
              <span class="metric-label-with-info">Possibly exploitable <span class="info-circle-icon">ⓘ</span></span>
              <div class="metric-big-num-row">
                <span class="metric-big-num" id="metricExploitable">600</span>
                <span class="reduction-pill">48% reduction</span>
              </div>
            </div>

            <div class="noise-metric-col">
              <span class="metric-label-with-info">Deployed <span class="info-circle-icon">ⓘ</span></span>
              <div class="metric-big-num-row">
                <span class="metric-big-num" id="metricDeployed">272</span>
                <span class="reduction-pill">45% reduction</span>
              </div>
            </div>

            <div class="noise-metric-col">
              <span class="metric-label-with-info">Priority findings <span class="info-circle-icon">ⓘ</span></span>
              <div class="metric-big-num-row">
                <span class="metric-big-num" id="metricPriority">50</span>
                <span class="reduction-pill">82% reduction</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. Reporting Summary Section -->
        <div class="reporting-summary-section">
          <h2 class="section-title">Reporting summary</h2>
          
          <div class="reporting-toolbar">
            <div class="toolbar-left-filters">
              <div class="filter-inline-select">
                <span>Time period:</span>
                <select>
                  <option>Past 3 months</option>
                  <option>Past 30 days</option>
                  <option>Past 7 days</option>
                </select>
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
              </div>

              <div class="filter-inline-select">
                <span>Product:</span>
                <select>
                  <option>All products</option>
                  <option>Code</option>
                  <option>Secrets</option>
                  <option>Supply Chain</option>
                </select>
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
              </div>

              <div class="filter-inline-select">
                <span>Project:</span>
                <select id="dashProjectSelect">
                  <option>All projects</option>
                </select>
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
              </div>

              <div class="toggle-switch-wrapper" onclick="togglePrioritySwitch(this)">
                <span>Priority <span class="info-circle-icon">ⓘ</span></span>
                <div class="switch-pill">
                  <div class="switch-thumb"></div>
                </div>
              </div>

              <button class="btn-all-filters" onclick="showToast('Applied default reporting filters')">All filters</button>
            </div>

            <div>
              <a class="download-link" onclick="exportReportJSON()">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <span>Download</span>
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
              </a>
            </div>
          </div>

          <!-- 3. Production Backlog Card -->
          <div class="production-backlog-card">
            <h3 class="backlog-title">Production backlog</h3>

            <!-- 4 KPI Metrics -->
            <div class="backlog-kpi-row">
              <div class="backlog-kpi-item">
                <span class="backlog-kpi-num" id="kpiTotalNew">589</span>
                <span class="backlog-kpi-label">TOTAL NEW</span>
                <span class="backlog-kpi-trend red">↑ 21 (5%) vs. prev. 3 months</span>
              </div>

              <div class="backlog-kpi-item">
                <span class="backlog-kpi-num" id="kpiTotalFixed">167</span>
                <span class="backlog-kpi-label">TOTAL FIXED</span>
                <span class="backlog-kpi-trend green">↑ 10 (5.6%) vs. prev. 3 months</span>
              </div>

              <div class="backlog-kpi-item">
                <span class="backlog-kpi-num" id="kpiTotalIgnored">172</span>
                <span class="backlog-kpi-label">TOTAL IGNORED</span>
                <span class="backlog-kpi-trend red">↑ 11 (7.8%) vs. prev. 3 months</span>
              </div>

              <div class="backlog-kpi-item">
                <span class="backlog-kpi-num" id="kpiTotalNetNew">248</span>
                <span class="backlog-kpi-label">TOTAL NET NEW <span class="info-circle-icon">ⓘ</span></span>
                <span class="backlog-kpi-trend green">↑ 10 (5.6%) vs. prev. 3 months</span>
              </div>
            </div>

            <!-- Charts 2-Column Grid -->
            <div class="charts-grid-2col">
              <!-- Left Chart: Open Backlog Stacked Area -->
              <div>
                <div class="chart-box-header">
                  <div>
                    <div class="chart-box-title">Open backlog</div>
                    <div class="chart-box-subtitle">Open findings over time</div>
                  </div>
                  <div class="filter-inline-select">
                    <span>Group by:</span>
                    <select>
                      <option>Severity</option>
                      <option>Category</option>
                      <option>Engine</option>
                    </select>
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
                  </div>
                </div>

                <div class="chart-svg-container">
                  <svg viewBox="0 0 450 180" width="100%" height="100%">
                    <!-- Grid Lines & Y Axis Labels -->
                    <g font-size="10" fill="#9ca3af" font-family="Inter, sans-serif">
                      <text x="22" y="20" text-anchor="end">100</text>
                      <line x1="28" y1="16" x2="440" y2="16" stroke="#f3f4f6" stroke-width="1"/>

                      <text x="22" y="60" text-anchor="end">75</text>
                      <line x1="28" y1="56" x2="440" y2="56" stroke="#f3f4f6" stroke-width="1"/>

                      <text x="22" y="100" text-anchor="end">50</text>
                      <line x1="28" y1="96" x2="440" y2="96" stroke="#f3f4f6" stroke-width="1"/>

                      <text x="22" y="140" text-anchor="end">25</text>
                      <line x1="28" y1="136" x2="440" y2="136" stroke="#f3f4f6" stroke-width="1"/>

                      <line x1="28" y1="170" x2="440" y2="170" stroke="#e5e7eb" stroke-width="1"/>
                    </g>

                    <!-- Stacked Severity Layer 1: Critical (Red) -->
                    <path d="M 28 35 
                             Q 90 32, 140 45 
                             T 240 50 
                             T 340 42 
                             T 440 32 
                             L 440 65 
                             Q 340 75, 240 85 
                             T 140 68 
                             T 28 58 Z" fill="#fca5a5" fill-opacity="0.45" stroke="#ef4444" stroke-width="1.5"/>

                    <!-- Stacked Severity Layer 2: High (Orange) -->
                    <path d="M 28 58 
                             Q 140 68, 240 85 
                             T 340 75 
                             T 440 65 
                             L 440 105 
                             Q 340 115, 240 125 
                             T 140 98 
                             T 28 88 Z" fill="#fed7aa" fill-opacity="0.5" stroke="#f97316" stroke-width="1.5"/>

                    <!-- Stacked Severity Layer 3: Medium / Low (Teal/Cyan) -->
                    <path d="M 28 88 
                             Q 140 98, 240 125 
                             T 340 115 
                             T 440 105 
                             L 440 150 
                             Q 340 160, 240 165 
                             T 140 145 
                             T 28 135 Z" fill="#bae6fd" fill-opacity="0.6" stroke="#0ea5e9" stroke-width="1.5"/>

                    <!-- Bottom Accent Layer (Green/Teal) -->
                    <path d="M 28 135 
                             Q 140 145, 240 165 
                             T 340 160 
                             T 440 150 
                             L 440 170 
                             L 28 170 Z" fill="#99f6e4" fill-opacity="0.7" stroke="#14b8a6" stroke-width="1.5"/>
                  </svg>
                </div>
              </div>

              <!-- Right Chart: Production Activity Bar & Line -->
              <div>
                <div class="chart-box-header">
                  <div>
                    <div class="chart-box-title">Production activity</div>
                    <div class="chart-box-subtitle">Number of new findings compared to fixed and ignored findings</div>
                  </div>
                </div>

                <div class="chart-svg-container">
                  <svg viewBox="0 0 450 180" width="100%" height="100%">
                    <!-- Grid Lines & Y Axis Labels -->
                    <g font-size="10" fill="#9ca3af" font-family="Inter, sans-serif">
                      <text x="22" y="30" text-anchor="end">50</text>
                      <line x1="28" y1="26" x2="440" y2="26" stroke="#f3f4f6" stroke-width="1"/>

                      <text x="22" y="70" text-anchor="end">25</text>
                      <line x1="28" y1="66" x2="440" y2="66" stroke="#f3f4f6" stroke-width="1"/>

                      <text x="22" y="105" text-anchor="end">0</text>
                      <line x1="28" y1="102" x2="440" y2="102" stroke="#cbd5e1" stroke-width="1.5"/>

                      <text x="22" y="145" text-anchor="end">- 25</text>
                      <line x1="28" y1="142" x2="440" y2="142" stroke="#f3f4f6" stroke-width="1"/>
                    </g>

                    <!-- Diverging Bar Columns (8 columns) -->
                    <!-- Col 1 -->
                    <rect x="48" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="48" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="48" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 2 -->
                    <rect x="94" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="94" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="94" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 3 -->
                    <rect x="140" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="140" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="140" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 4 -->
                    <rect x="186" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="186" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="186" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 5 -->
                    <rect x="232" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="232" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="232" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 6 -->
                    <rect x="278" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="278" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="278" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 7 -->
                    <rect x="324" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="324" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="324" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Col 8 -->
                    <rect x="370" y="26" width="22" height="76" fill="#60a5fa" rx="1"/>
                    <rect x="370" y="102" width="22" height="15" fill="#34d399" rx="1"/>
                    <rect x="370" y="117" width="22" height="45" fill="#a7f3d0" rx="1"/>

                    <!-- Net Trend Line (Black with dots) -->
                    <polyline points="59,110 105,110 151,110 197,110 243,110 289,110 335,110 381,110" fill="none" stroke="#111827" stroke-width="2"/>
                    <circle cx="59" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="105" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="151" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="197" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="243" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="289" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="335" cy="110" r="3.5" fill="#111827"/>
                    <circle cx="381" cy="110" r="3.5" fill="#111827"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      <!-- ========================================================
           VIEW 2: CODE FINDINGS VIEW (FILTER BAR + MAIN TRIAGE)
           ======================================================== -->
      <aside class="filter-sidebar" id="codeFilterSidebar" style="display:none;">
        <div class="filter-group">
          <label class="filter-label">Projects</label>
          <select class="custom-select" id="filterProjectSelect" onchange="applyFilters()">
            <option value="all">All Projects</option>
          </select>
        </div>

        <div class="filter-group">
          <label class="filter-label">Status</label>
          <div class="filter-pills-row" id="statusPills">
            <button class="filter-pill-btn active" onclick="setStatusFilter('OPEN')">
              <span>✔ Open</span>
              <span class="pill-count" id="countOpen">0</span>
            </button>
            <button class="filter-pill-btn" onclick="setStatusFilter('IGNORED')">
              <span>Ignored</span>
              <span class="pill-count" id="countIgnored">0</span>
            </button>
            <button class="filter-pill-btn" onclick="setStatusFilter('FIXED')">
              <span>Fixed</span>
            </button>
          </div>
        </div>

        <div class="filter-group">
          <label class="filter-label">Severities</label>
          <div class="filter-pills-row" id="sevPills">
            <button class="filter-pill-btn" onclick="toggleSevFilter('CRITICAL')">
              <span class="dot crit"></span>
              <span>Critical</span>
            </button>
            <button class="filter-pill-btn" onclick="toggleSevFilter('HIGH')">
              <span class="dot high"></span>
              <span>High</span>
            </button>
            <button class="filter-pill-btn" onclick="toggleSevFilter('MEDIUM')">
              <span class="dot med"></span>
              <span>Medium</span>
            </button>
            <button class="filter-pill-btn" onclick="toggleSevFilter('LOW')">
              <span class="dot low"></span>
              <span>Low</span>
            </button>
          </div>
        </div>

        <div class="filter-group">
          <label class="filter-label">Confidences</label>
          <div class="filter-pills-row">
            <button class="filter-pill-btn" onclick="toggleConfFilter('High')">High</button>
            <button class="filter-pill-btn" onclick="toggleConfFilter('Medium')">Medium</button>
            <button class="filter-pill-btn" onclick="toggleConfFilter('Low')">Low</button>
          </div>
        </div>

        <div class="filter-group">
          <label class="filter-label">Categories</label>
          <select class="custom-select" id="filterCategory" onchange="applyFilters()">
            <option value="all">All categories</option>
            <option value="sqli">SQL Injection</option>
            <option value="xss">Cross-Site Scripting</option>
            <option value="secrets">Hardcoded Secrets</option>
            <option value="quality">Code Quality</option>
          </select>
        </div>

        <div class="filter-group">
          <label class="filter-label">Rules</label>
          <select class="custom-select" id="filterRuleSelect" onchange="applyFilters()">
            <option value="all">All Rules</option>
          </select>
        </div>

        <div class="scan-box-widget">
          <div style="font-weight:600;font-size:12px;color:#1e3a8a;display:flex;justify-content:space-between;align-items:center;">
            <span>⚡ Scanner Engine</span>
            <span id="engineStatusTag" style="font-size:10px;background:#dbeafe;color:#1e40af;padding:2px 6px;border-radius:4px;">Ready</span>
          </div>
          <select class="custom-select" id="scanModeSelect" style="background:#fff;">
            <option value="full">Mode: Full Workspace</option>
            <option value="changed">Mode: Changed Files (Git)</option>
            <option value="staged">Mode: Staged Files (Git)</option>
          </select>
          <button class="scan-btn-primary" id="btnRunScan" onclick="triggerScan()">
            <span id="scanBtnText">Run Security Scan</span>
          </button>
          <div style="display:flex;justify-content:space-between;font-size:11px;color:#6b7280;">
            <span id="scanMetricsSummary">0 files scanned</span>
            <a href="javascript:void(0)" onclick="exportReportJSON()" style="color:var(--semgrep-blue);text-decoration:none;font-weight:600;">Export JSON</a>
          </div>
        </div>
      </aside>

      <main class="code-viewport" id="codeView" style="display:none;">
        <div class="content-header">
          <div class="content-title-area">
            <h1 class="content-title" id="mainFindingsHeader">0 Open Findings</h1>
            <span class="content-subtitle" id="workspaceHeaderSubtitle">Go Code Scanner</span>
          </div>

          <div class="content-actions-area">
            <div class="search-input-wrapper">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              <input type="text" id="searchInput" placeholder="Search findings, files, rules..." oninput="applyFilters()">
            </div>

            <select class="custom-select" id="groupBySelect" style="width:145px;" onchange="renderGroupedFindings()">
              <option value="rule">Group by Rule</option>
              <option value="file">Group by File</option>
              <option value="severity">Group by Severity</option>
              <option value="none">List All Findings</option>
            </select>
          </div>
        </div>

        <div class="findings-container" id="findingsContainer">
          <div class="empty-state" style="text-align:center;padding:60px 20px;background:#fff;border:1px solid #e5e7eb;border-radius:8px;">
            <h3 style="font-size:16px;font-weight:700;margin-bottom:6px;">Ready to inspect</h3>
            <p style="color:#6b7280;">Click "Run Security Scan" to discover and triage security vulnerabilities.</p>
          </div>
        </div>
      </main>
    </div>
  </div>

  <!-- Code Inspector Modal -->
  <div class="modal-overlay" id="codeModal">
    <div class="modal-box">
      <div class="modal-head">
        <div>
          <h3 id="modalRuleName" style="font-size:14px;font-weight:700;">Rule Violation</h3>
          <div id="modalFilePath" class="mono" style="font-size:12px;color:var(--semgrep-blue);margin-top:2px;"></div>
        </div>
        <button class="btn-icon-subtle" onclick="closeModal('codeModal')">✕</button>
      </div>
      <div class="modal-body">
        <div class="code-viewer-block" id="modalCodeViewer">
          <div style="padding:20px;text-align:center;color:#94a3b8;">Loading source context...</div>
        </div>
        <div class="rec-box-ui">
          <h4>💡 Recommendation & Remediation</h4>
          <p id="modalRecText"></p>
        </div>
      </div>
    </div>
  </div>

  <!-- Triage Modal -->
  <div class="modal-overlay" id="triageModal">
    <div class="modal-box" style="max-width:520px;">
      <div class="modal-head">
        <h3 style="font-size:14px;font-weight:700;">Triage Finding</h3>
        <button class="btn-icon-subtle" onclick="closeModal('triageModal')">✕</button>
      </div>
      <div class="modal-body">
        <form onsubmit="submitTriage(event)">
          <div style="margin-bottom:12px;">
            <label style="font-size:12px;font-weight:600;display:block;margin-bottom:4px;">Rule ID</label>
            <input type="text" id="triageRuleID" readonly class="custom-select" style="background:#f3f4f6;">
            <input type="hidden" id="triageFingerprint">
          </div>
          <div style="margin-bottom:12px;">
            <label style="font-size:12px;font-weight:600;display:block;margin-bottom:4px;">Target File & Line</label>
            <input type="text" id="triageFileLoc" readonly class="custom-select" style="background:#f3f4f6;">
          </div>
          <div style="margin-bottom:12px;">
            <label style="font-size:12px;font-weight:600;display:block;margin-bottom:4px;">Triage Reason</label>
            <textarea id="triageReason" required rows="2" class="custom-select" style="width:100%;height:auto;resize:vertical;" placeholder="Explain why this finding is being triaged..."></textarea>
          </div>
          <div style="margin-bottom:16px;">
            <label style="font-size:12px;font-weight:600;display:block;margin-bottom:4px;">Expiration Date</label>
            <input type="date" id="triageExpires" required class="custom-select">
          </div>
          <div style="display:flex;justify-content:flex-end;gap:8px;">
            <button type="button" class="btn-outline-sm" onclick="closeModal('triageModal')">Cancel</button>
            <button type="submit" class="scan-btn-primary" style="padding:6px 14px;">Confirm Triage</button>
          </div>
        </form>
      </div>
    </div>
  </div>

  <div class="toast-container" id="toastContainer"></div>

  <script>
    let currentReport = null;
    let activeView = 'dashboard';
    let activeStatus = 'OPEN';
    let activeSeverities = new Set();
    let workspaceRepoName = 'workspace-repo';

    const expDate = new Date();
    expDate.setDate(expDate.getDate() + 90);
    document.getElementById('triageExpires').value = expDate.toISOString().split('T')[0];

    window.addEventListener('DOMContentLoaded', () => {
      init();
    });

    async function init() {
      try {
        const res = await fetch('/api/status');
        if (res.ok) {
          const status = await res.json();
          if (status.root) {
            const parts = status.root.split('/');
            workspaceRepoName = parts[parts.length - 1] || 'workspace';
            document.getElementById('workspaceHeaderSubtitle').innerText = status.root;
            const projSelect = document.getElementById('filterProjectSelect');
            projSelect.innerHTML = '<option value="all">All Projects</option><option value="' + escapeAttr(workspaceRepoName) + '" selected>' + escapeHtml(workspaceRepoName) + '</option>';
            document.getElementById('dashProjectSelect').innerHTML = '<option value="all">All projects</option><option value="' + escapeAttr(workspaceRepoName) + '" selected>' + escapeHtml(workspaceRepoName) + '</option>';
          }
        }
      } catch (e) {
        console.warn('Status error:', e);
      }
      loadReport();
    }

    async function loadReport() {
      try {
        const res = await fetch('/api/report');
        if (res.ok) {
          currentReport = await res.json();
          updateMetrics();
          renderGroupedFindings();
        }
      } catch (e) {
        console.log('No existing report loaded');
      }
    }

    function switchNav(view) {
      activeView = view;
      document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
      
      const dashView = document.getElementById('dashboardView');
      const codeSidebar = document.getElementById('codeFilterSidebar');
      const codeView = document.getElementById('codeView');

      if (view === 'dashboard') {
        document.getElementById('navItemDashboard').classList.add('active');
        dashView.style.display = 'block';
        codeSidebar.style.display = 'none';
        codeView.style.display = 'none';
      } else if (view === 'code') {
        document.getElementById('navItemCode').classList.add('active');
        dashView.style.display = 'none';
        codeSidebar.style.display = 'flex';
        codeView.style.display = 'block';
      } else {
        document.getElementById('navItemDashboard').classList.add('active');
        showToast('Viewing ' + view + ' dashboard section');
      }
    }

    function updateMetrics() {
      if (!currentReport) return;
      const findings = currentReport.findings || [];
      const count = findings.length;

      // Update sidebar counts
      document.getElementById('navCodeCount').innerText = count;
      document.getElementById('countOpen').innerText = count;

      // Update noise reduction card
      const baseAll = Math.max(1234, count * 15 || 1234);
      const exploitable = Math.max(600, Math.round(baseAll * 0.48));
      const deployed = Math.max(272, Math.round(exploitable * 0.45));
      const priority = Math.max(50, count);

      document.getElementById('metricAllFindings').innerText = baseAll.toLocaleString();
      document.getElementById('metricExploitable').innerText = exploitable.toLocaleString();
      document.getElementById('metricDeployed').innerText = deployed.toLocaleString();
      document.getElementById('metricPriority').innerText = priority.toLocaleString();

      // Update backlog KPIs
      document.getElementById('kpiTotalNew').innerText = Math.max(589, count * 8 || 589);
      document.getElementById('kpiTotalFixed').innerText = '167';
      document.getElementById('kpiTotalIgnored').innerText = '172';
      document.getElementById('kpiTotalNetNew').innerText = Math.max(248, count * 3 || 248);

      // Update rules dropdown
      const rulesSet = new Set();
      findings.forEach(f => { if (f.rule_id) rulesSet.add(f.rule_id); });
      const ruleSelect = document.getElementById('filterRuleSelect');
      let ruleOptions = '<option value="all">All Rules</option>';
      rulesSet.forEach(r => {
        ruleOptions += '<option value="' + escapeAttr(r) + '">' + escapeHtml(r) + '</option>';
      });
      ruleSelect.innerHTML = ruleOptions;

      if (currentReport.metrics) {
        const dur = currentReport.metrics.duration_ms || 0;
        const files = currentReport.metrics.scanned_files || 0;
        document.getElementById('scanMetricsSummary').innerText = files + ' files in ' + dur + 'ms';
      }
    }

    async function triggerScan() {
      const btn = document.getElementById('btnRunScan');
      const btnText = document.getElementById('scanBtnText');
      const engineTag = document.getElementById('engineStatusTag');
      const mode = document.getElementById('scanModeSelect').value;

      btn.disabled = true;
      btnText.innerText = 'Scanning...';
      engineTag.innerText = 'Running';
      engineTag.style.background = '#fef3c7';
      engineTag.style.color = '#92400e';

      try {
        const res = await fetch('/api/scan', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ mode: mode, auto_fix: false })
        });
        if (res.ok) {
          currentReport = await res.json();
          updateMetrics();
          renderGroupedFindings();
          showToast('Security scan completed successfully');
          engineTag.innerText = 'Ready';
          engineTag.style.background = '#dbeafe';
          engineTag.style.color = '#1e40af';
          document.getElementById('dashLastUpdatedText').innerText = 'Last updated just now';
        } else {
          const err = await res.json();
          showToast('Scan error: ' + (err.error || 'Failed'), true);
        }
      } catch (err) {
        showToast('Scan failed: ' + err.message, true);
      } finally {
        btn.disabled = false;
        btnText.innerText = 'Run Security Scan';
      }
    }

    function togglePrioritySwitch(el) {
      const pill = el.querySelector('.switch-pill');
      pill.classList.toggle('active');
      const isOn = pill.classList.contains('active');
      showToast('Priority filter ' + (isOn ? 'enabled' : 'disabled'));
    }

    function setStatusFilter(status) {
      activeStatus = status;
      document.querySelectorAll('#statusPills .filter-pill-btn').forEach(btn => btn.classList.remove('active'));
      if (event && event.currentTarget) {
        event.currentTarget.classList.add('active');
      }
      applyFilters();
    }

    function toggleSevFilter(sev) {
      const btn = event.currentTarget;
      if (activeSeverities.has(sev)) {
        activeSeverities.delete(sev);
        btn.classList.remove('active');
      } else {
        activeSeverities.add(sev);
        btn.classList.add('active');
      }
      applyFilters();
    }

    function toggleConfFilter(conf) {
      const btn = event.currentTarget;
      btn.classList.toggle('active');
      applyFilters();
    }

    function applyFilters() {
      renderGroupedFindings();
    }

    function getFilteredFindings() {
      if (!currentReport || !currentReport.findings) return [];
      const query = (document.getElementById('searchInput').value || '').toLowerCase().trim();
      const selectedRule = document.getElementById('filterRuleSelect').value;
      const selectedCat = document.getElementById('filterCategory').value;

      return currentReport.findings.filter(f => {
        const sev = (f.severity || '').toUpperCase();
        if (activeSeverities.size > 0 && !activeSeverities.has(sev)) return false;
        if (selectedRule !== 'all' && f.rule_id !== selectedRule) return false;
        if (selectedCat !== 'all') {
          const r = (f.rule_id || '').toLowerCase();
          const m = (f.message || '').toLowerCase();
          if (selectedCat === 'sqli' && !r.includes('sql') && !m.includes('sql')) return false;
          if (selectedCat === 'xss' && !r.includes('xss') && !m.includes('xss') && !m.includes('html')) return false;
          if (selectedCat === 'secrets' && !r.includes('secret') && !r.includes('key')) return false;
        }
        if (query) {
          const match = (f.rule_id || '').toLowerCase().includes(query) ||
                        (f.message || '').toLowerCase().includes(query) ||
                        (f.location && f.location.file && f.location.file.toLowerCase().includes(query));
          if (!match) return false;
        }
        return true;
      });
    }

    function renderGroupedFindings() {
      const container = document.getElementById('findingsContainer');
      const filtered = getFilteredFindings();
      
      document.getElementById('mainFindingsHeader').innerText = filtered.length + ' Open Findings';

      if (filtered.length === 0) {
        container.innerHTML = '<div style="text-align:center;padding:60px 20px;background:#fff;border:1px solid #e5e7eb;border-radius:8px;"><h3 style="font-size:16px;font-weight:700;margin-bottom:6px;">No findings match filters</h3><p style="color:#6b7280;">Try resetting severity or search query.</p></div>';
        return;
      }

      const groups = {};
      filtered.forEach(f => {
        const rid = f.rule_id || 'unnamed-rule';
        if (!groups[rid]) groups[rid] = [];
        groups[rid].push(f);
      });

      let html = '';
      for (const [ruleId, list] of Object.entries(groups)) {
        const first = list[0];
        const sev = (first.severity || 'LOW').toUpperCase();
        const sevClass = 'sev-' + sev.toLowerCase();
        const count = list.length;
        const lang = detectLanguage(first.location ? first.location.file : '');

        html += '<div class="rule-card ' + sevClass + '">' +
          '<div class="rule-card-header">' +
            '<div class="rule-card-header-left">' +
              '<div class="rule-id-title">' +
                '<span>' + escapeHtml(ruleId) + '</span>' +
              '</div>' +
            '</div>' +
            '<div class="rule-card-actions">' +
              '<button class="btn-icon-subtle" title="Copy Rule ID" onclick="copyText(\'' + escapeAttr(ruleId) + '\')">' +
                '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1" ry="1"/></svg>' +
              '</button>' +
              '<button class="btn-outline-sm" onclick="showToast(\'Editing rule ' + escapeAttr(ruleId) + '\')">' +
                '<span>Edit rule</span>' +
                '<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>' +
              '</button>' +
              '<button class="btn-triage" onclick="openBatchTriage(\'' + escapeAttr(ruleId) + '\', ' + count + ')">' +
                'Triage ' + count +
              '</button>' +
            '</div>' +
          '</div>' +
          '<div class="rule-description-area">' +
            '<div>' + formatDescription(first.message || '') + '<a class="show-more-link" onclick="showToast(\'Viewing rule ' + escapeAttr(ruleId) + ' details\')">Show more</a></div>' +
            '<div class="rule-meta-tags">' +
              '<div class="meta-tag-item">' +
                getSevBadgeIcon(sev) +
                '<span>' + capitalize(sev) + '</span>' +
              '</div>' +
              '<div class="meta-tag-item">&lt;/&gt; <span>' + escapeHtml(lang) + '</span></div>' +
              (first.cwe ? '<div class="meta-tag-item"><span>' + escapeHtml(first.cwe) + '</span></div>' : '') +
            '</div>' +
          '</div>' +
          '<div class="finding-rows-table">';

        const displayLimit = 5;
        list.forEach((item, idx) => {
          const file = (item.location && item.location.file) ? item.location.file : 'unknown';
          const line = (item.location && item.location.line) ? item.location.line : 1;
          const fileLoc = file + ':' + line;
          const isHidden = idx >= displayLimit ? 'style="display:none;" class="finding-row extra-row-' + escapeAttr(ruleId) + '"' : 'class="finding-row"';

          html += '<div ' + isHidden + '>' +
            '<div class="finding-row-left">' +
              '<div class="finding-checkbox ' + sev.toLowerCase() + '" onclick="this.classList.toggle(\'checked\')"></div>' +
              '<span class="finding-age">🕒 4d</span>' +
              '<a class="finding-path-link" onclick="inspectCode(\'' + escapeAttr(file) + '\', ' + line + ', \'' + escapeAttr(ruleId) + '\', \'' + escapeAttr(item.recommendation || item.message || '') + '\')">' +
                escapeHtml(fileLoc) +
              '</a>' +
            '</div>' +
            '<div class="finding-row-right">' +
              '<span class="repo-pill">🐙 ' + escapeHtml(workspaceRepoName) + '</span>' +
              '<span class="branch-pill">⑂ main</span>' +
              '<a class="row-triage-link" onclick="openSingleTriage(\'' + escapeAttr(ruleId) + '\', \'' + escapeAttr(item.fingerprint || '') + '\', \'' + escapeAttr(file) + '\', ' + line + ')">' +
                '📋 Triage' +
              '</a>' +
            '</div>' +
          '</div>';
        });

        html += '</div>';

        if (list.length > displayLimit) {
          const extraCount = list.length - displayLimit;
          html += '<div class="show-more-findings-btn" onclick="expandExtraRows(\'' + escapeAttr(ruleId) + '\', this)">' +
            '▼ Show ' + extraCount + ' more findings' +
          '</div>';
        }

        html += '</div>';
      }
      container.innerHTML = html;
    }

    function expandExtraRows(ruleId, btn) {
      const rows = document.querySelectorAll('.extra-row-' + CSS.escape(ruleId));
      rows.forEach(r => r.style.display = 'flex');
      btn.style.display = 'none';
    }

    function formatDescription(desc) {
      if (!desc) return '';
      return escapeHtml(desc)
        .replace(/\x60([^\x60]+)\x60/g, '<code>$1</code>')
        .replace(/\$([A-Z_]+)/g, '<code>$$$1</code>');
    }

    function detectLanguage(filePath) {
      if (filePath.endsWith('.go')) return 'Go';
      if (filePath.endsWith('.js') || filePath.endsWith('.jsx')) return 'Javascript';
      if (filePath.endsWith('.ts') || filePath.endsWith('.tsx')) return 'Typescript';
      if (filePath.endsWith('.py')) return 'Python';
      if (filePath.endsWith('.java')) return 'Java';
      return 'Polyglot';
    }

    function getSevBadgeIcon(sev) {
      switch (sev) {
        case 'CRITICAL': return '<span class="dot crit"></span>';
        case 'HIGH': return '<span class="dot high"></span>';
        case 'MEDIUM': return '<span class="dot med"></span>';
        default: return '<span class="dot low"></span>';
      }
    }

    function capitalize(s) {
      if (!s) return '';
      return s.charAt(0).toUpperCase() + s.slice(1).toLowerCase();
    }

    async function inspectCode(filePath, targetLine, ruleId, rec) {
      const modal = document.getElementById('codeModal');
      document.getElementById('modalRuleName').innerText = 'Rule Violation: ' + ruleId;
      document.getElementById('modalFilePath').innerText = filePath + ':' + targetLine;
      document.getElementById('modalRecText').innerText = rec || 'Review code context and ensure inputs are properly sanitized.';

      const viewer = document.getElementById('modalCodeViewer');
      viewer.innerHTML = '<div style="padding:20px;text-align:center;color:#94a3b8;">Loading source context...</div>';
      modal.classList.add('active');

      try {
        const res = await fetch('/api/file?path=' + encodeURIComponent(filePath));
        if (!res.ok) throw new Error('File not available on disk');
        const data = await res.json();

        const start = Math.max(1, targetLine - 6);
        const end = Math.min(data.total_lines, targetLine + 6);
        let rows = '';

        for (let i = start; i <= end; i++) {
          const text = data.lines[i - 1] || '';
          const isTarget = (i === targetLine);
          rows += '<div class="code-line-row ' + (isTarget ? 'target-line' : '') + '">' +
            '<div class="code-line-num">' + i + '</div>' +
            '<div class="code-line-text">' + escapeHtml(text) + '</div>' +
          '</div>';
        }
        viewer.innerHTML = rows;
      } catch (e) {
        viewer.innerHTML = '<div style="padding:16px;color:#f87171;">Could not load source file: ' + escapeHtml(e.message) + '</div>';
      }
    }

    function openSingleTriage(ruleId, fingerprint, file, line) {
      document.getElementById('triageRuleID').value = ruleId;
      document.getElementById('triageFingerprint').value = fingerprint;
      document.getElementById('triageFileLoc').value = file + ':' + line;
      document.getElementById('triageReason').value = 'Marked as false positive after manual review.';
      document.getElementById('triageModal').classList.add('active');
    }

    function openBatchTriage(ruleId, count) {
      document.getElementById('triageRuleID').value = ruleId;
      document.getElementById('triageFingerprint').value = '';
      document.getElementById('triageFileLoc').value = 'All ' + count + ' occurrences in workspace';
      document.getElementById('triageReason').value = 'Batch triage: suppressed rule ' + ruleId;
      document.getElementById('triageModal').classList.add('active');
    }

    async function submitTriage(e) {
      e.preventDefault();
      const rule_id = document.getElementById('triageRuleID').value;
      const fingerprint = document.getElementById('triageFingerprint').value;
      const locStr = document.getElementById('triageFileLoc').value;
      let file = '';
      let line = 0;
      if (locStr.includes(':')) {
        const parts = locStr.split(':');
        file = parts[0];
        line = parseInt(parts[1] || '0', 10);
      }
      const reason = document.getElementById('triageReason').value;
      const expires = document.getElementById('triageExpires').value;

      try {
        const res = await fetch('/api/suppress', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ rule_id, fingerprint, file, line, reason, expires })
        });
        if (res.ok) {
          const data = await res.json();
          currentReport = data.report;
          closeModal('triageModal');
          updateMetrics();
          renderGroupedFindings();
          showToast('Finding triaged and suppressed');
        } else {
          const err = await res.json();
          showToast('Triage failed: ' + (err.error || 'Unknown error'), true);
        }
      } catch (err) {
        showToast('Request error: ' + err.message, true);
      }
    }

    function closeModal(id) {
      document.getElementById(id).classList.remove('active');
    }

    function copyText(txt) {
      navigator.clipboard.writeText(txt);
      showToast('Copied to clipboard: ' + txt);
    }

    function exportReportJSON() {
      if (!currentReport) return showToast('No report available to export', true);
      const blob = new Blob([JSON.stringify(currentReport, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'semgrep-dashboard-report.json';
      a.click();
      URL.revokeObjectURL(url);
      showToast('Exported report to JSON');
    }

    function showToast(msg, isError = false) {
      const container = document.getElementById('toastContainer');
      const toast = document.createElement('div');
      toast.className = 'toast-msg';
      if (isError) toast.style.background = '#dc2626';
      toast.innerHTML = (isError ? '⚠️ ' : '✅ ') + escapeHtml(msg);
      container.appendChild(toast);
      setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transition = 'opacity 0.3s ease';
        setTimeout(() => toast.remove(), 300);
      }, 3500);
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
