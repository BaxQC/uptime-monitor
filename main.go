package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// ── Config ────────────────────────────────────────────────────────────────────

type Site struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Sites     []Site `json:"sites"`
	ChannelID string `json:"channel_id"`
}

func loadSites() Config {
	data, err := os.ReadFile("sites.json")
	if err != nil {
		log.Fatal("Cannot read sites.json:", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatal("Invalid sites.json:", err)
	}
	return cfg
}

// ── Checker ───────────────────────────────────────────────────────────────────

type CheckResult struct {
	Site        Site
	Up          bool
	StatusCode  int
	ResponseMs  int64
	Error       string
	CheckedAt   time.Time
}

func checkSite(site Site) CheckResult {
	start := time.Now()
	result := CheckResult{
		Site:      site,
		CheckedAt: start,
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(site.URL)
	result.ResponseMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Up = false
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Up = resp.StatusCode >= 200 && resp.StatusCode < 400
	if !result.Up {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return result
}

// ── Discord ───────────────────────────────────────────────────────────────────

func responseColor(up bool, ms int64) int {
	if !up {
		return 0xED4245 // red
	}
	if ms > 2000 {
		return 0xFAA61A // yellow — slow
	}
	return 0x57F287 // green
}

func responseEmoji(up bool, ms int64) string {
	if !up {
		return "🔴"
	}
	if ms > 2000 {
		return "🟡"
	}
	return "🟢"
}

func speedLabel(ms int64) string {
	switch {
	case ms < 300:
		return "Fast"
	case ms < 1000:
		return "Normal"
	case ms < 2000:
		return "Slow"
	default:
		return "Very Slow"
	}
}

func buildEmbed(results []CheckResult) *discordgo.MessageEmbed {
	allUp := true
	for _, r := range results {
		if !r.Up {
			allUp = false
			break
		}
	}

	var fields []*discordgo.MessageEmbedField
	for _, r := range results {
		emoji := responseEmoji(r.Up, r.ResponseMs)

		var value string
		if r.Up {
			value = fmt.Sprintf(
				"**Status:** `%d OK`\n**Response:** `%dms` — %s",
				r.StatusCode, r.ResponseMs, speedLabel(r.ResponseMs),
			)
		} else {
			value = fmt.Sprintf(
				"**Status:** `DOWN`\n**Error:** `%s`",
				r.Error,
			)
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s %s", emoji, r.Site.Name),
			Value: value,
		})
	}

	title := "✅ All Systems Operational"
	color := 0x57F287
	if !allUp {
		title = "🚨 Outage Detected"
		color = 0xED4245
	}

	checkedAt := results[0].CheckedAt.UTC().Format("Jan 2, 2006 at 15:04:05 UTC")

	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  color,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Checked at %s • every 5 minutes", checkedAt),
		},
	}
}

func sendOrEdit(s *discordgo.Session, channelID string, msgID *string, embed *discordgo.MessageEmbed) {
	if *msgID != "" {
		// Try to edit the existing message
		_, err := s.ChannelMessageEditEmbed(channelID, *msgID, embed)
		if err == nil {
			return
		}
		// If edit fails (message deleted etc.), fall through to send new
		log.Printf("Edit failed, sending new message: %v", err)
	}

	msg, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Failed to send Discord message: %v", err)
		return
	}
	*msgID = msg.ID
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN not set in .env")
	}

	cfg := loadSites()
	if cfg.ChannelID == "" {
		log.Fatal("channel_id not set in sites.json")
	}
	if len(cfg.Sites) == 0 {
		log.Fatal("No sites defined in sites.json")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	if err := dg.Open(); err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	log.Printf("✅ Uptime monitor started — watching %d site(s)", len(cfg.Sites))
	log.Printf("📢 Posting to channel: %s", cfg.ChannelID)

	// Track the last posted message so we edit it instead of spamming
	msgID := ""

	run := func() {
		log.Println("⏱  Checking sites...")
		var results []CheckResult
		for _, site := range cfg.Sites {
			r := checkSite(site)
			if r.Up {
				log.Printf("  🟢 %s — %dms", site.Name, r.ResponseMs)
			} else {
				log.Printf("  🔴 %s — DOWN: %s", site.Name, r.Error)
			}
			results = append(results, r)
		}
		embed := buildEmbed(results)
		sendOrEdit(dg, cfg.ChannelID, &msgID, embed)
	}

	// Run immediately on start, then every 5 minutes
	run()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		run()
	}
}
