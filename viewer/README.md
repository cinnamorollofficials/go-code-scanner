# Security Finding Dashboard (`viewer/`)

Standalone Interactive Security Dashboard & Scan Runner for **go-code-scanner**.

## Features

- **⚡ Real-time Scan Execution**: Trigger scans directly from the browser UI with customizable options (Full scan, Changed files, Staged files, profile selection, auto-fix).
- **📊 Executive Security Metrics**: Live severity breakdown, donut distribution, scan duration, and file counters.
- **🔍 Deep Code Inspector**: Inspect vulnerable source code lines in context with line numbers and violation highlights.
- **🛡️ 1-Click Suppression**: Form to quickly add suppressions (`--reason`, `--expires`) with instant rescan.
- **💾 Export**: Download JSON report or print/PDF view.

## Quick Start

### Method 1: Using the Go CLI (Zero setup required)
```bash
go run ./cmd/security-review ui --port 8080
```
Open [http://localhost:8080](http://localhost:8080) in your browser.

### Method 2: Frontend Dev Server (with Hot Reload)
```bash
cd viewer
npm install
npm run dev
```
Open [http://localhost:5173](http://localhost:5173) in your browser. Make sure the Go backend is running on `:8080`.
