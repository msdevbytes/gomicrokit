# 🧰 GoMicroKit — Scaffold Go Microservices Fast with Style 🚀

**GoMicroKit** is an interactive CLI tool for generating scalable, idiomatic Go microservice boilerplates — with support for REST frameworks, databases, GORM, Docker, and more. Inspired by Laravel's
`artisan`, but for Go developers.

✨ Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Cobra](https://github.com/spf13/cobra), and Go best practices.

---

## 📦 Features

- 🏗️ Generate a full microservice directory with:
  - Framework: `fiber` (more coming soon)
  - DB setup `mysql` (more coming soon)
  - GORM support toggle
  - Docker support
- ⚙️ Uses Repository Pattern, Service Layer, DTOs
- 🧪 Auto-generates test boilerplates
- 🐳 Dockerfile support
- 🌈 Interactive UI with Bubble Tea (fancy CLI)
- 🧽 Remove previously generated services
- 🕹️ Flags & prompts: hybrid control for automation and UX

---

## 🔧 Installation

```bash
go install github.com/msdevbytes/gomicrokit@latest
```

> Or clone and run locally:

```bash
git clone https://github.com/msdevbytes/gomicrokit.git
cd gomicrokit
go run main.go
```

---

## 🚀 Usage

### 🎬 Start the CLI

```bash
gomicrokit new
```

You'll be guided through:

- Project name
- Framework selection (currently supports `fiber`)
- Database selection
- Use GORM? (y/n)
- Include Dockerfile? (y/n)

✨ A spinner will run while your project is generated, packages are installed, and success is displayed in style.

---

## 🧪 Example

```bash
gomicrokit new
```

```
🧱 Project Name:      mysvc
🛠️  Framework:         fiber
🗄️  Database:          postgres
📦 Use GORM:          y
🐳 Docker Support:    y
```

➡️ Output:

```bash
📁 Scaffolding...
✅ Created: internal/service/mysvc_service.go
✅ Created: internal/repository/mysvc_repository.go
✅ Created: internal/model/mysvc_model.go
✅ Created: internal/handler/mysvc_handler.go
✅ Updated: internal/service/container.go
✅ Updated: internal/routes/index.go
📦 Installing packages...
🔥 Done! Project 'mysvc' generated successfully.
```

---

## 📁 Project Structure

```bash
mysvc/
├── cmd/
│   └── main.go
├── internal/
│   ├── dto/
│   ├── handler/
│   ├── model/
│   ├── repository/
│   ├── routes/
│   └── service/
├── test/
│   └── unit/
│       └── dto/
├── go.mod
├── Dockerfile (optional)
└── .gen_history.json
```

---

## ⚡ Commands

### 🏗️ Generate a Service (non-interactive)

```bash
gomicrokit make:service --name=event --force
```

### 🧼 Remove a Service (interactive)

```bash
gomicrokit remove:service
```

---

## ✨ Templates

Templates are stored in:

```
templates/service/
├── model.tmpl
├── repository.tmpl
├── service.tmpl
├── handler.tmpl
├── dto.tmpl
├── dto_test.tmpl
```

---

## 📖 Dev Guide

```bash
go run main.go
```

---

## 📄 License

MIT — use it freely.

---

## 💬 Credits

Built with ❤️ by [msdevbytes](https://github.com/msdevbytes).
