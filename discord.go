package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

const (
	ebuyerAvatar = "https://media.glassdoor.com/sqll/764749/ebuyer-squarelogo-1396431281679.png"
)

func (b *bot) startUp(num int) error {
	hook := discordgo.WebhookParams{
		Content: fmt.Sprintf("GPU stock bot is live! Watching %d GPU's.", num),
	}
	reqBody, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	_, err = http.Post(b.DiscordURL, "application/json", bytes.NewBuffer(reqBody))
	return err
}

func (b *bot) sendDiscord(gpu, url, model string) error {
	embed := &discordgo.MessageEmbed{
		URL:    url,
		Author: &discordgo.MessageEmbedAuthor{Name: "eBuyer", IconURL: ebuyerAvatar, ProxyIconURL: ebuyerAvatar, URL: url},
		Title:  gpu,
		Color:  39219,
	}

	hook := discordgo.WebhookParams{
		Content: fmt.Sprintf("%s now in stock!", model),
		Embeds:  []*discordgo.MessageEmbed{embed},
	}
	reqBody, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	_, err = http.Post(b.DiscordURL, "application/json", bytes.NewBuffer(reqBody))
	return err
}

func (b *bot) outOfStock(gpu, url, model string) error {
	embed := &discordgo.MessageEmbed{
		URL:    url,
		Author: &discordgo.MessageEmbedAuthor{Name: "eBuyer", IconURL: ebuyerAvatar, ProxyIconURL: ebuyerAvatar, URL: url},
		Title:  gpu,
		Color:  13369344,
	}

	hook := discordgo.WebhookParams{
		Content: fmt.Sprintf("%s now out of stock :rage:", model),
		Embeds:  []*discordgo.MessageEmbed{embed},
	}
	reqBody, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	_, err = http.Post(b.DiscordURL, "application/json", bytes.NewBuffer(reqBody))
	return err
}
