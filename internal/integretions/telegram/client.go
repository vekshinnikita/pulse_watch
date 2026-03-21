package telegram_integration

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	Bot *tgbotapi.BotAPI
}

func NewDefaultClient() (*Client, error) {
	cfg := GetConfig()

	bot, err := tgbotapi.NewBotAPI(cfg.ApiToken)
	if err != nil {
		return nil, err
	}

	return &Client{Bot: bot}, nil
}

func (c *Client) SendMessage(chatId int, text string) error {
	msg := tgbotapi.NewMessage(int64(chatId), text)
	_, err := c.Bot.Send(msg)

	return err
}
