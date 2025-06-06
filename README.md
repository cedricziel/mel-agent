# 🤖 MEL Agent

## The AI-First Workflow Automation Platform

**MEL Agent** is an open-source platform for building, deploying, and managing AI-powered workflows. Think of it as **n8n meets GPT** – a visual workflow builder specifically designed for the AI era.

## 🌟 Why MEL Agent?

### **AI-Native Design**

Unlike traditional automation platforms retrofitted for AI, MEL Agent is built from the ground up for LLM orchestration and AI agent workflows.

### **Visual Workflow Builder**

- **Drag-and-drop interface** inspired by n8n, Make.com, and Node-RED
- **Real-time debugging** with step-by-step execution traces
- **Live data preview** for every node in your workflow
- **Built-in error handling** and retry policies

### **Enterprise Ready** _(Future Opportunities)_

- **Multi-tenant architecture** with row-level security _(Opportunity for SaaS providers)_
- **SOC2 Type II compliance** _(Opportunity for enterprise adoption)_
- **GDPR compliant** with automatic data lineage _(Opportunity for EU markets)_
- **High availability** for production workloads _(Opportunity for critical systems)_

### **Developer Experience**

- **API-first design** – everything in the UI is available via REST API
- **Infrastructure as Code** support _(Opportunity for GitOps workflows)_
- **Extensive node library** for common AI and automation tasks
- **Custom node development** with Go SDK

## 🚀 Key Features

### **Powerful Node Types**

- **🧠 AI Nodes**: LLM chat, embeddings _(opportunity for image generation, speech-to-text)_
- **🔗 Integration Nodes**: Slack, Baserow, webhooks, HTTP requests _(opportunity for Gmail, Notion, etc.)_
- **⚡ Logic Nodes**: If/else, transformations, delays, switches _(opportunity for advanced loops)_
- **📊 Data Nodes**: Database queries, variable management _(opportunity for file operations)_
- **🔧 Utility Nodes**: Logging, merging, splitting, workflow calls

### **Advanced Data Flow**

- **Envelope-based architecture** for robust data propagation
- **Automatic splitting and merging** of array data _(opportunity for advanced parallel processing)_
- **Global variables** and context sharing across nodes
- **Binary attachments** support _(opportunity for multimedia workflows)_

### **Real-Time Collaboration**

- **Live editing** with WebSocket-based synchronization
- **Version control** with semantic versioning
- **Execution history** with full replay capability
- **Team sharing** and permission management _(opportunity for enterprise collaboration)_

## 🏗️ Architecture

MEL Agent follows modern cloud-native patterns:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  React Frontend │    │   Go Backend     │    │   PostgreSQL    │
│                 │◄──►│                  │◄──►│                 │
│ • Visual Builder│    │ • REST API       │    │ • Multi-tenant  │
│ • Real-time UI  │    │ • WebSocket      │    │ • Encrypted     │
│ • Debug Tools   │    │ • Node Engine    │    │ • JSONB Storage │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

**Technology Stack:**

- **Backend**: Go with Chi router, PostgreSQL, WebSockets
- **Frontend**: React, Vite, Tailwind CSS, ReactFlow
- **Infrastructure**: Docker, Docker Compose, Kubernetes-ready

## 🏃‍♂️ Quick Start

### Option 1: Docker Compose (Recommended)

```bash
git clone https://github.com/cedricziel/mel-agent.git
cd mel-agent
docker compose up --build
```

**🎉 That's it!** Your MEL Agent instance is running:

- **API**: <http://localhost:8080>
- **Web UI**: <http://localhost:5173>
- **Health Check**: <http://localhost:8080/health>

### Option 2: Development Setup

**Prerequisites**: Go 1.21+, Node.js 18+, PostgreSQL 15+

```bash
# Backend
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/melagent?sslmode=disable"
go run ./cmd/server

# Frontend
cd web
pnpm install && pnpm dev
```

## 📖 Documentation

### **Getting Started**

- [🏗️ Architecture Overview](docs/design/0-agents.md)
- [🔌 Connections & Integrations](docs/design/1-connections.md)
- [🎨 Visual Builder Guide](docs/design/2-builder.md)
- [📊 Data Flow & Envelopes](docs/design/3-data-flow.md)

### **Development** _(Contribution Opportunities)_

- [🧩 Building Custom Nodes](docs/development/custom-nodes.md) _(opportunity to contribute documentation)_
- [🔧 API Reference](docs/api/README.md) _(opportunity to generate from code)_
- [🚀 Deployment Guide](docs/deployment/README.md) _(opportunity for K8s, cloud guides)_

### **Examples** _(Community Opportunities)_

- [📧 Email Processing Agent](examples/email-agent/) _(opportunity for real-world examples)_
- [📱 Social Media Manager](examples/social-agent/) _(opportunity for marketing automation)_
- [📊 Data Pipeline Automation](examples/data-agent/) _(opportunity for ETL workflows)_

## 🌍 Heritage & Inspiration

MEL Agent stands on the shoulders of giants, drawing inspiration from:

- **[n8n](https://n8n.io/)** – Visual workflow automation and node-based architecture
- **[Make.com](https://make.com/)** – Intuitive drag-and-drop interface and robust integrations  
- **[Node-RED](https://nodered.org/)** – Flow-based programming concepts and real-time debugging
- **[Zapier](https://zapier.com/)** – Ease of use and extensive integration ecosystem

**What makes MEL Agent different:**

- ✨ **AI-first design** – Built specifically for LLM and AI agent workflows
- 🏗️ **Modern architecture** – Go backend, React frontend, envelope-based data flow
- 🔧 **Developer-friendly** – Comprehensive APIs, custom node SDK _(opportunity for IaC support)_
- 🏢 **Open Source** – Community-driven development, no vendor lock-in

## 🤝 Contributing

We welcome contributions! MEL Agent is built by the community, for the community.

- **🐛 Report bugs** via GitHub Issues
- **💡 Suggest features** through GitHub Discussions
- **🔧 Submit PRs** – see CONTRIBUTING.md for guidelines _(opportunity to establish contribution guidelines)_

### **Development Commands**

```bash
# Backend
go test ./...              # Run tests
go vet ./...              # Lint
go build ./cmd/server     # Build

# Frontend
cd web
pnpm lint                 # Lint
pnpm build                # Build
pnpm test                 # Test (coming soon)
```

## 📜 License

MEL Agent is open source software licensed under the [MIT License](LICENSE).

## 🙏 Acknowledgments

Special thanks to the open source community and the teams behind n8n, Make.com, Node-RED, and Zapier for pioneering visual workflow automation.

---

**Ready to build your first AI agent?** 🚀

[Get started in 5 minutes →](docs/getting-started.md) _(opportunity for quick-start guide)_
