package handlers

import (
	"bot/bot"
	"log"
	"smoke3/db"
)

type MenuHandler struct {
}

func (t *MenuHandler) Handle(c *bot.Context) *bot.Response {
	log.Println("MenuHandler START")
	c.NextHandler = nil
	groups, err := db.GetGroupsByAccount(toDomainAccount(c.BotAccount))
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Меню"
	if len(groups) > 0 {
		r.AddButtonString("Группы", &GroupsHandler{groups})
	}
	r.AddButtonString("Создать группу", &CreateGroupHandler{})
	r.AddButtonString("Назад", &StartHandler{})
	log.Println("MenuHandler END")
	return r
}
