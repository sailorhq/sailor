# Sailor

<p align="center">
    <img width="100" src="https://github.com/codekidX/sailor/blob/main/assets/sailor.png?raw=true">
</p>

---

Sailor is an open source system for managing application configs and secrets for
your backend applications. It acts as a single source of truth for all your
configs and maintains it under a single umbrella for ease of management. It
helps you validate (using rules), deploy, hot-reload and rollback your configs
and secrets.

Sailor can be setup quickly and the [sailor client]() is responsible for
fetching, maintaining and updating your configs, which takes most of the
boilerplate code from your hands. It is built to be frugal in resource usage.

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
