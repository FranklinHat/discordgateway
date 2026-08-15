package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type PingPayload struct {
	Source string `json:"source"`
}

type PingResponse struct {
	Status     string `json:"status"`
	Reply      string `json:"reply"`
	EchoSource string `json:"echo_source"`
}

type NotifyPayload struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Timestamp   time.Time           `json:"timestamp,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
}

type DiscordWebhookPayload struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type AlertPayload struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Host    string `json:"host"`
}

func getColorForStatus(status string) int {
	switch status {
	case "critical", "down", "error":
		return 15158332 // red
	case "warning", "degraded":
		return 15844367 // yellow
	case "up", "healthy", "ok":
		return 3066993 // green
	case "info":
		return 3447003 // blue
	default:
		return 9807270 // grey
	}
}

func sendToDiscord(webhookURL string, payload DiscordWebhookPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("could not marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned status: %d (%s)", resp.StatusCode, resp.Status)
	}
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	res := HealthResponse{
		Status:  "ok",
		Message: "Gateway operational.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload PingPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := PingResponse{
		Status:     "ok",
		Reply:      "pong!!",
		EchoSource: payload.Source,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func NotifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload NotifyPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		http.Error(w, "DISCORD_WEBHOOK_URL not set", http.StatusBadRequest)
		return
	}

	formattedContent := fmt.Sprintf("[%s] %s", payload.Sender, payload.Message)
	msg := DiscordWebhookPayload{
		Content: formattedContent,
	}

	if err := sendToDiscord(webhookURL, msg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to relay message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "notification successfully relayed",
	})
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload AlertPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		http.Error(w, "DISCORD_WEBHOOK_URL not set", http.StatusBadRequest)
		return
	}

	embed := DiscordEmbed{
		Title:       fmt.Sprintf("Alert Notification: %s is %s", payload.Service, payload.Status),
		Description: payload.Message,
		Color:       getColorForStatus(payload.Status),
		Timestamp:   time.Now(),
		Fields: []DiscordEmbedField{
			{
				Name:   "Host",
				Value:  payload.Host,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  payload.Status,
				Inline: true,
			},
		},
	}

	discordPayload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	if err := sendToDiscord(webhookURL, discordPayload); err != nil {
		http.Error(w, fmt.Sprintf("Failed to relay message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "notification successfully relayed",
	})
}

func rawDiscordProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		http.Error(w, "DISCORD_WEBHOOK_URL not set", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/notify", NotifyHandler)
	mux.HandleFunc("/alert", alertHandler)
	mux.HandleFunc("/raw", rawDiscordProxyHandler)

	log.Println("Starting server on :8089...")
	if err := http.ListenAndServe(":8089", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
