package util

import (
	"bot/bot"
	"smoke3/domain"
)

func ToBotAccount(domainAcc *domain.Account) *bot.BotAccount {
	return &bot.BotAccount{
		FirstName: domainAcc.FirstName,
		LastName:  domainAcc.LastName,
		ChatId:    domainAcc.ChatId,
	}
}