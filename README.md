# 🐧 Sailor

> Configuration Management for Cloud Native Workloads

<p align="center">
    <img width="200" src="https://github.com/codekidX/sailor/blob/main/assets/sailor.png?raw=true">
</p>

Modern cloud-native applications often drown in configuration sprawl. **Sailor**
helps you take the helm with a developer-friendly service to manage, audit, and
apply configurations for your containerized applications — effortlessly.

**Sailor** is a configuration management service built for developers running
containerized applications inside Kubernetes. It simplifies, secures, and
streamlines how configuration is handled across environments and teams.

## 🚀 Features

- 🌊 Centralized config management
- 🔐 Secure secrets and policy enforcement
- 🔁 Versioning & rollback
- ⚙️ Easy integration with your CI/CD or GitOps setup
- 🧠 Developer-friendly UI & API

## 🧭 Quick Start

```bash
# Download Sailor Binary
curl -sL https://getsailor.sh | bash

# Run Sailor
./sailor

[🐧] Running on port :7766
admin: http://localhost:7766/_admin
```

---

### Roadmap

| Feature              | Support |
| -------------------- | ------- |
| Validation Layer     | ✅      |
| State Fallback       | ⚠️      |
| Control Center {UI}  | ⚠️      |
| Access-Control / IAM | ❌      |
| Auditing             | ❌      |

### Supported Clients

| Language | Support |
| -------- | ------- |
| Go       | ✅      |
| NodeJs   | ⚠️      |
