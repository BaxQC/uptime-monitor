<div align="center">

# 📡 Uptime Monitor

**Watches your URLs every 5 minutes and posts a live status embed in Discord.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00acd7?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Discord](https://img.shields.io/badge/Discord-Bot-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.com/developers)
[![Interval](https://img.shields.io/badge/Interval-5%20minutes-faa61a?style=for-the-badge)]()

</div>

---

## 💡 What it does

Every 5 minutes, the bot checks all your URLs and **edits a single embed** in your Discord channel — no spam, just one live status card that updates in place.

```
✅ All Systems Operational

🟢 My Website       Status: 200 OK  |  Response: 142ms — Fast
🟢 My API           Status: 200 OK  |  Response: 87ms  — Fast
🟡 Some Service     Status: 200 OK  |  Response: 1800ms — Slow

Checked at May 22, 2026 at 14:35:00 UTC • every 5 minutes
```

- 🟢 **Green** — up, response under 1s
- 🟡 **Yellow** — up, but slow (over 2s)
- 🔴 **Red** — down or HTTP error

---

## ⚙️ Setup

### 1. Clone & install

```bash
git clone https://github.com/BaxQC/uptime-monitor
cd uptime-monitor
go mod tidy
```

### 2. Configure your sites

Edit `sites.json`:

```json
{
  "channel_id": "123456789012345678",
  "sites": [
    { "name": "My Website", "url": "https://example.com" },
    { "name": "My API",     "url": "https://api.example.com/health" }
  ]
}
```

### 3. Set your bot token

```bash
cp .env.example .env
```

```env
DISCORD_TOKEN=your_bot_token_here
```

### 4. Run

```bash
go run main.go
# or build:
go build -o uptime-monitor .
./uptime-monitor
```

---

## 🤖 Discord Bot Setup

1. [discord.com/developers/applications](https://discord.com/developers/applications) → New Application
2. **Bot** → Reset Token → copy to `.env`
3. OAuth2 → URL Generator → scopes: `bot` → permissions: `Send Messages` + `Read Message History`
4. Invite bot to your server

**Getting Channel ID:**
- Developer Mode on → right-click the channel → **Copy Channel ID**

---

## 🗺️ Roadmap

- [ ] `@mention` role on outage
- [ ] Per-site custom check interval
- [ ] Response time graph over time
- [ ] HTTP POST / custom headers support (for authenticated endpoints)

---

## 📄 License

MIT © [Bax](https://github.com/BaxQC)
