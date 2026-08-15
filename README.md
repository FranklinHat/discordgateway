
A lightweight, high-performance webhook adapter and proxy written in Go. It centralizes homelab notifications, formats structured alerts into rich Discord embed cards, and isolates your real Discord webhook tokens behind a private microservice.


- **Transparent Discord Proxy (`/discord`):** Drop-in replacement endpoint for pre-built homelab tools (Uptime Kuma, Watchtower, etc.) without altering payload structures.
- **Dynamic Alert Cards (`/alert`):** Converts structured JSON into color-coded Discord embeds based on status level (`critical`, `warning`, `healthy`, `info`).
- **Simple Message Relay (`/notify`):** Lightweight sender/message notification relay.
- **Ultra-Lean Container:** Multi-stage Docker build resulting in a minimal Alpine runtime image (~15MB).
- **Secure Secret Isolation:** Keeps Discord webhook tokens out of multiple individual container configurations and scripts.

---

## 🛠️ Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Liveness check returning gateway status |
| `POST` | `/ping` | Test endpoint echoing request source |
| `POST` | `/notify` | Relays basic sender/message strings to Discord |
| `POST` | `/alert` | Formats homelab alerts into rich Discord Embeds |
| `POST` | `/discord` | Transparent proxy forwarding raw Discord payloads |

---

Work in progress
