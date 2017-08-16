package handlers

import (
	"strings"
	"log"
	"bot/bot"
	"smoke3/db"
	"smoke3/util"
)

type StartJoinGroupHandler struct {
}

func (t *StartJoinGroupHandler) Handle(c *bot.Context) *bot.Response {
	log.Println("StartJoinGroupHandler START")
	uuid := strings.Replace(c.Message.Text, "/start ", "", 1)
	g, err := db.GetGroupByUUID(uuid)
	if err != nil {
		return &bot.Response{
			Text: "Группа не найдена",
		}
	}

	if err := db.AddAccountToGroup(toDomainAccount(c.BotAccount), g); err != nil {
		if err == db.NotUnique {
			return &bot.Response{
				Text: "Вы уже состоите в группе *" + g.Name + "*",
			}
		}
		panic(err)
	}

	log.Println("StartJoinGroupHandler END")

	go func() {
		for _, acc := range g.Accounts {
			if acc.ChatId != c.BotAccount.ChatId {
				accContext := bot.GetContext(util.ToBotAccount(acc))
				r := &bot.Response{
					Text: "*" + c.BotAccount.FirstName + " " + c.BotAccount.LastName + "* присоеденился " +
						"в группу *" + g.Name + "*",
				}
				go accContext.Send(r)
			}
		}
	}()
	return &bot.Response{
		Text: "Добро Пожаловать в группу *" + g.Name + "*",
	}
}
