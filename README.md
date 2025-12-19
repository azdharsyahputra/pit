# PIT

**PIT** is a local development engine designed to keep developers focused on what matters most: **writing code**.

Inspired by a racing pit crew, PIT handles the surrounding infrastructure — web server, runtimes, tools, and system setup — so you don’t have to.

> You build the application.  
> **PIT handles the pit stop.**

---

## 🚀 What is PIT?

PIT is a **portable, engine-driven local development environment** built in Go.

It provides a fast, predictable, and zero-friction workflow by generating and managing everything required to run modern web applications locally — without manual configuration.

PIT is opinionated where it matters and invisible where it should be.

---

## ⚡ Core Features (v0.1)

### 🧠 Engine-First Architecture
- Single binary engine written in Go
- Centralized service orchestration
- Explicit, idempotent one-time system setup
- No manual config editing

### 🌐 Web Stack
- Portable Nginx
- Automatic virtual host generation
- Automatic `/etc/hosts` synchronization
- Bind privileged ports without running the engine as root

### 🐘 PHP Runtime
- Dedicated PHP-FPM per project
- Dedicated PHP-FPM runtime for tools
- UNIX socket communication (no random ports)
- Clear separation between application runtime and tooling

### 🧰 Built-in Tools
- phpMyAdmin included
- Tools run in isolated runtimes
- Tool virtual hosts are auto-generated and disposable

### 🗄️ Database Experience
- Optimized local database configuration
- Password-based root access for development
- phpMyAdmin works out-of-the-box

---

## 🧩 Why PIT?

PIT is built on a simple belief:

> Local development should feel **instant, predictable, and effortless**.

PIT takes ownership of infrastructure so developers don’t have to:
- No environment drift
- No hidden state
- No manual wiring
- No “read the docs first” moments

Everything is generated, managed, and controlled by the engine.

---

## 🗂️ Project Structure (Simplified)

```
pit/
├─ cmd/pit/                 # CLI entrypoint
├─ internal/
│  ├─ core/                 # Engine & orchestration
│  ├─ services/             # Service lifecycle
│  └─ tools/                # Tool management & vhost generation
├─ nginx/                   # Portable nginx
├─ runtime/
│  ├─ _tools/php/           # PHP-FPM runtime for tools
│  └─ <project>/php/        # PHP-FPM per project
├─ tools/
│  └─ phpmyadmin/           # Built-in tools
└─ config/
```

---

## 🚦 Getting Started

### 1️⃣ One-time setup
```bash
./pit setup
```

This will:
- Prepare system permissions
- Configure local database access

### 2️⃣ Start the engine
```bash
./pit start
```

### 3️⃣ Open tools
```
http://phpmyadmin.test
```

Login:
- **User:** `root`
- **Password:** *(empty)*

---

## 📌 Project Status

- **Version:** `v0.1.0`
- **Stage:** Stable MVP
- **Scope:** Local development only

PIT v0.1 focuses on **core stability and developer experience**.  
Future versions will expand only after real-world usage.

---

## 🛣️ Roadmap (Post v0.1)

Planned directions:
- Service status & health reporting
- Tool lifecycle management
- Lightweight control panel
- Cross-platform packaging
- Additional built-in tools

---

## ⚠️ Disclaimer

PIT is intended for **local development only**.

System defaults prioritize speed and clarity over production-grade hardening.

---

## 🏁 Philosophy

In a race, drivers don’t stop to adjust their engine or tires.

They trust the pit crew.

**PIT is that crew for your local development.**

---

## 📄 License

MIT License.