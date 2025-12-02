package identity

import identityv1 "github.com/gaming-platform/api/go/identity/v1"

type Bot struct {
	BotId    string
	Username string
}

func fromProtoBot(protoBot *identityv1.Bot) *Bot {
	if protoBot == nil {
		return nil
	}

	return &Bot{
		BotId:    protoBot.BotId,
		Username: protoBot.Username,
	}
}
