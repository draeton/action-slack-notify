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
	EnvSlackChannel     = "SLACK_CHANNEL"
	EnvSlackFooter      = "SLACK_FOOTER"
	EnvSlackIconEmoji   = "SLACK_ICON_EMOJI"
	EnvSlackLinkNames   = "SLACK_LINK_NAMES"
	EnvSlackMessage     = "SLACK_MESSAGE"
	EnvSlackMessageLink = "SLACK_MESSAGE_LINK"
	EnvSlackUserName    = "SLACK_USERNAME"
	EnvSlackWebhook     = "SLACK_WEBHOOK"
)

type Webhook struct {
	AsUser      bool    `json:"as_user"`
	Blocks      []Block `json:"blocks,omitempty"`
	Channel     string  `json:"channel,omitempty"`
	IconEmoji   string  `json:"icon_emoji,omitempty"`
	IconURL     string  `json:"icon_url,omitempty"`
	LinkNames   string  `json:"link_names,omitempty"`
	UnfurlLinks bool    `json:"unfurl_links"`
	UserName    string  `json:"username,omitempty"`
}

type Block struct {
	Type      string    `json:"type,omitempty"`
	Text      string    `json:"text,omitempty"`
	Accessory Accessory `json:"accessory,omitempty"`
}

type Accessory struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
}

func main() {
	endpoint := os.Getenv(EnvSlackWebhook)
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "URL is required")
		os.Exit(1)
	}
	text := os.Getenv(EnvSlackMessage)
	if text == "" {
		fmt.Fprintln(os.Stderr, "Message is required")
		os.Exit(1)
	}
	if strings.HasPrefix(os.Getenv("GITHUB_WORKFLOW"), ".github") {
		os.Setenv("GITHUB_WORKFLOW", "Link to action run")
	}

	blocks := []Block{
		{
			Type: "section",
			Text: text,
		},
	}

	link := envOr(EnvSlackMessageLink, "")

	if link != "" {
		blocks[0].Accessory = Accessory{
			Type: "button",
			Text: "view",
			Url:  link,
		}
	}

	footer := envOr(EnvSlackFooter, "<https://github.com/rtCamp/github-actions-library|Powered By rtCamp's GitHub Actions Library>")

	if footer != "" {
		footerBlocks := []Block{
			{
				Type: "divider",
			}, {
				Type: "section",
				Text: footer,
			},
		}
		blocks = append(blocks, footerBlocks...)
	}

	msg := Webhook{
		AsUser:    false,
		UserName:  os.Getenv(EnvSlackUserName),
		IconEmoji: os.Getenv(EnvSlackIconEmoji),
		Channel:   os.Getenv(EnvSlackChannel),
		LinkNames: os.Getenv(EnvSlackLinkNames),
		Blocks:    blocks,
	}

	if err := send(endpoint, msg); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %s\n", err)
		os.Exit(2)
	}
}

func envOr(name, def string) string {
	if d, ok := os.LookupEnv(name); ok {
		return d
	}
	return def
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
		return fmt.Errorf("Error on message: %s\n%s\n", res.Status, json.NewEncoder(os.Stdout).Encode(msg))
	}
	fmt.Println(res.Status)
	return nil
}
