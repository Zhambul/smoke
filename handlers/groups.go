package handlers

import (
	"bot/bot"
	"smoke3/db"
	"smoke3/domain"
	"log"
	"smoke3/util"
)

type GroupsHandler struct {
	groups []*domain.Group
}

func (h *GroupsHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.Text = "Группы"
	r.ClearButtons()

	for _, g := range h.groups {
		r.AddButtonString(g.Name, &OneGroupHandler{group: g, groups: h.groups})
	}
	r.AddButtonString("Назад", &MenuHandler{})
	return r
}

type CreateGroupHandler struct {
}

func (h *CreateGroupHandler) Handle(c *bot.Context) *bot.Response {
	c.NextHandler = &GiveGroupNameHandler{}

	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Дай название группе"
	r.AddButtonString("Назад", &MenuHandler{})
	return r
}

type GiveGroupNameHandler struct {
}

func (h *GiveGroupNameHandler) Handle(c *bot.Context) *bot.Response {
	log.Println("GiveGroupNameHandler START")
	groupName := c.Message.Text
	g, err := db.CreateNewGroup(toDomainAccount(c.BotAccount), groupName)
	if err != nil {
		if err == db.NotUnique {
			c.NextHandler = h
			return &bot.Response{
				Text: "Группа с таким названием уже есть, попробуйте снова",
			}
		} else {
			panic(err)
		}
	}

	r := &bot.Response{
		Text: "Группа *" + groupName + "* создана",
	}

	r.AddButton(util.ShareButton(g))
	r.AddButtonString("Назад", &MenuHandler{})
	log.Printf("GiveGroupNameHandler END + %+v\n", r)
	return r
}

type OneGroupHandler struct {
	group  *domain.Group
	groups []*domain.Group
}

func (h *OneGroupHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.ClearButtons()
	res := "*" + h.group.Name + "*:\n"
	if len(h.group.Accounts) == 1 {
		res += "В группе кроме вас никого"
	} else {
		for _, acc := range h.group.Accounts {
			res += acc.FirstName + " " + acc.LastName + "\n"
		}
	}
	r.Text = res

	r.AddButton(util.ShareButton(h.group))

	if h.group.CreatorAccount.ChatId != c.BotAccount.ChatId {
		r.AddButtonString("Покинуть", &LeaveGroupHandle{h.group})
	}
	if h.group.CreatorAccount.ChatId == c.BotAccount.ChatId {
		r.AddButtonString("Удалить", &DeleteGroupHandler{h.group})
	}
	r.AddButtonString("Назад", &GroupsHandler{h.groups})

	return r
}

type LeaveGroupHandle struct {
	group *domain.Group
}

func (h *LeaveGroupHandle) Handle(c *bot.Context) *bot.Response {
	err := db.LeaveGroup(h.group, toDomainAccount(c.BotAccount))
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}
	r := c.CurrentResponse
	r.Text = "Вы покинули группу *" + h.group.Name + "*"
	r.ClearButtons()
	r.AddButtonString("Меню", &MenuHandler{})
	return r
}

type DeleteGroupHandler struct {
	group *domain.Group
}

func (h *DeleteGroupHandler) Handle(c *bot.Context) *bot.Response {
	err := db.DeleteGroup(h.group)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}
	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Группа *" + h.group.Name + "* удалена"
	r.AddButtonString("Меню", &MenuHandler{})
	return r
}
