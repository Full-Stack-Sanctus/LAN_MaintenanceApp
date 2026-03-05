# LAN Maintenance App

A cross-platform **LAN monitoring tool** built with a modern desktop stack. This application combines **web technologies**, **native system access**, and **high-performance Go networking logic** to deliver fast and reliable local network diagnostics.

---

## 🏗 Architecture Overview

```
Vue (Frontend UI)
        ↓
Rust (Tauri Core Bridge)
        ↓
Go Binary (Network Engine)
        ↓
Operating System (SNMP, ARP, Network Stack)
```

### 1️⃣ Frontend — Tauri v2 + Vue + TypeScript

* **Framework:** Tauri v2
* **UI Template:** Vue
* **Language:** TypeScript / JavaScript
* **Package Manager:** npm

The frontend provides a modern, reactive UI built with Vue and TypeScript.
Tauri enables packaging the web UI as a **native desktop application** for:

* Windows (.exe)
* macOS (.app)
* Linux

Tauri bridges web technologies and native system capabilities while maintaining a small bundle size and improved security compared to traditional Electron apps.

---

### 2️⃣ Bridge Layer — Rust + Tauri Sidecar

The bridge layer is responsible for secure communication between the UI and the Go backend.

#### 🔹 Tauri Sidecar

Tauri’s **Sidecar** feature allows the app to:

* Run a compiled Go binary as a child process
* Communicate via standard input/output
* Exchange structured JSON data

#### 🔹 Rust Bridge Logic

Rust acts as the middleman:

1. Vue requests a network scan.
2. Rust receives the request via Tauri command.
3. Rust executes the Go binary (sidecar).
4. Go performs the network operations.
5. Go returns JSON output.
6. Rust forwards JSON to Vue.

This ensures:

* Strong type safety
* Process isolation
* Secure IPC communication

---

### 3️⃣ Backend — Go (Golang)

The backend is written in **Go** for performance and system-level networking access.

#### Responsibilities:

* SNMP scanning
* ARP table inspection
* Device discovery
* LAN diagnostics
* Network metadata extraction

Go is compiled separately for each target OS:

* Windows
* macOS
* Linux

Each compiled binary is bundled with Tauri during the build process.

---

## 🚀 Key Features

* Cross-platform desktop support
* High-performance LAN scanning
* SNMP device interrogation
* ARP-based device discovery
* Native system-level networking access
* Secure UI ↔ Backend communication
* Lightweight desktop packaging

---

## ⚙️ How It Works (Execution Flow)

1. User clicks **Scan Network**.
2. Vue triggers a Tauri command.
3. Rust executes the Go binary using Sidecar.
4. Go scans the LAN and generates JSON results.
5. Rust captures output.
6. Vue renders structured device data in the UI.

---

## 🛠 Build Process

### 1️⃣ Build Go Binary

```bash
cd src-go
go build -o ../src-tauri/binaries/network-engine
```

This produces the OS-specific binary used by Tauri.

---

### 2️⃣ Configure Tauri

Add the binary to `tauri.conf.json`:

```json
{
  "bundle": {
    "externalBin": ["binaries/network-engine"]
  }
}
```

---

### 3️⃣ Build Desktop App

```bash
npm install
npm run tauri build
```

This will generate:

* Windows `.exe`
* macOS `.app`
* Linux package

Privileges (Windows/macOS): 

* Unmanaged scanning (ARP) requires 
** Administrator (Windows) or  
** Root/Sudo (macOS/Linux) privileges because it uses raw network sockets. If the app is run as a standard user, the scanUnmanaged function will fail.

---

## 🔐 Why Tauri?

Tauri is a powerful desktop framework that:

* Bridges web tech with native system capabilities
* Uses the system WebView (smaller footprint than Electron)
* Enforces strict IPC security
* Allows execution of native binaries (like Go)

It leverages heavy lifting from the underlying OS while keeping the application lightweight and secure.

---

## 📦 Tech Stack Summary

| Layer      | Technology       |
| ---------- | ---------------- |
| UI         | Vue + TypeScript |
| Desktop    | Tauri v2         |
| Bridge     | Rust             |
| Backend    | Go (Golang)      |
| Networking | SNMP, ARP        |

---

## 📌 Future Enhancements

* Real-time monitoring
* Background network watcher
* Device history tracking
* Exportable reports
* Authentication layer
* Remote management module

---
