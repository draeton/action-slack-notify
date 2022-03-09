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
	EnvSlackLinks   = "SLACK_LINKS"
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
	Elements []Button `json:"elements,omitempty"`
	Text     *Text    `json:"text,omitempty"`
	Type     string   `json:"type"`
}

type Button struct {
	Type string `json:"type,omitempty"`
	Text *Text  `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
}

type Text struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type Link struct {
	Text string `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
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
	if strings.HasPrefix(os.Getenv("GITHUB_WORKFLOW"), ".github") {
		_ = os.Setenv("GITHUB_WORKFLOW", "Link to action run")
	}

	blocks := []Block{
		{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: text,
			},
		},
	}

	var links []Link
	var buttons []Button

	if str := os.Getenv(EnvSlackLinks); str != "" {
		if err := json.Unmarshal([]byte(str), &links); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error on parse: %s, `%s`\n", err, str)
			os.Exit(2)
		}

		if len(links) > 0 {
			for _, link := range links {
				buttons = append(buttons, Button{
					Type: "button",
					Text: &Text{
						Type: "plain_text",
						Text: link.Text,
					},
					Url: link.Url,
				})
			}
		}

		if len(buttons) > 0 {
			blocks = append(blocks,
				Block{
					Elements: buttons,
					Text:     nil,
					Type:     "actions",
				},
			)
		}
	}

	msg := Webhook{
		AsUser:  false,
		Blocks:  blocks,
		Channel: os.Getenv(EnvSlackChannel),
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
