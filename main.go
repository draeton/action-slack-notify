package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	EnvSlackChannel = "SLACK_CHANNEL"
	EnvSlackFooter  = "SLACK_FOOTER"
	EnvSlackMessage = "SLACK_MESSAGE"
	EnvSlackWebhook = "SLACK_WEBHOOK"
)

type Webhook struct {
	AsUser  bool    `json:"as_user,omitempty"`
	Blocks  []Block `json:"blocks,omitempty"`
	Channel string  `json:"channel,omitempty"`
	Text    string  `json:"text,omitempty"`
}

type Block struct {
	Elements *[]Text `json:"elements,omitempty"`
	Text     *Text   `json:"text,omitempty"`
	Type     string  `json:"type"`
}

type Text struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

func main() {
	endpoint := os.Getenv(EnvSlackWebhook)
	if endpoint == "" {
		_, _ = fmt.Fprintln(os.Stderr, "URL is required")
		os.Exit(1)
	}
	text := os.Getenv(EnvSlackMessage)
	if text == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Message is required")
		os.Exit(1)
	}
	footer := os.Getenv(EnvSlackFooter)
	if strings.HasPrefix(os.Getenv("GITHUB_WORKFLOW"), ".github") {
		_ = os.Setenv("GITHUB_WORKFLOW", "Link to action run")
	}

	blocks := []Block{
		{
			Text: &Text{
				Text: text,
				Type: "mrkdwn",
			},
			Type: "section",
		},
	}

	if footer != "" {
		blocks = append(blocks,
			Block{
				Text: nil,
				Type: "divider",
			},
			Block{
				Elements: &[]Text{
					{
						Text: footer,
						Type: "mrkdwn",
					},
				},
				Text: nil,
				Type: "context",
			},
		)
	}

	msg := Webhook{
		AsUser:  false,
		Blocks:  blocks,
		Channel: os.Getenv(EnvSlackChannel),
		Text:    text,
	}

	if err := send(endpoint, msg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error sending message: %s\n", err)
		os.Exit(2)
	}
}

func send(endpoint string, msg Webhook) error {
	enc, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(enc)
	res, err := http.Post(endpoint, "application/json", b)
	if err != nil {
		return err
	}

	if res.StatusCode >= 299 {
		data, _ := json.MarshalIndent(msg, "", "\t")
		return fmt.Errorf("Error on message: %s, `%s`\n", res.Status, data)
	}
	fmt.Println(res.Status)
	return nil
}
