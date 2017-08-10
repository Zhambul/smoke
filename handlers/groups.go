package handlers

import (
	"strings"
	"bot/bot"
	"smoke3/db"
	"smoke3/domain"
	"log"
)

type StartJoinGroupHandler struct {
}

func (t *StartJoinGroupHandler) Handle(c *bot.Context) *bot.Response {
	log.Println("StartJoinGroupHandler START")
	uuid := strings.Replace(c.Message.Text, "/start ", "", 1)
	g, err := db.GetGroupByUUID(uuid)
	if err != nil {
		panic(err)
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
	return &bot.Response{
		Text: "Добро Пожаловать",
	}
}

type StartHandler struct {
}

func (t *StartHandler) Handle(c *bot.Context) *bot.Response {
	r := &bot.Response{}
	r.Text = "Привет"

	r.AddButtonString("Го курить!", &GoSmokeHandler{})
	r.AddButtonString("Меню", &MenuHandler{})
	return r
}

type MenuHandler struct {
}

func (t *MenuHandler) Handle(c *bot.Context) *bot.Response {
	log.Println("MenuHandler START")
	groups, err := db.GetGroupsByAccount(toDomainAccount(c.BotAccount))
	if err != nil {
		panic(err)
	}

	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Меню"
	if len(groups) > 0 {
		r.AddButtonString("Группы", &GroupsHandler{groups})
	}
	r.AddButtonString("Создать группу", &CreateGroupHandler{})
	log.Println("MenuHandler END")
	return r
}

type GroupsHandler struct {
	groups []*domain.Group
}

func (h *GroupsHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.Text = "Группы"
	r.ClearButtons()

	for _, g := range h.groups {
		r.AddButtonString(g.Name, &OneGroupHandler{g})
	}
	return r
}

type CreateGroupHandler struct {
}

func (h *CreateGroupHandler) Handle(c *bot.Context) *bot.Response {
	c.NextHandler = &GiveGroupNameHandler{}

	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Дай название группе"
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

	r.AddButton(&bot.Button{
		Text:              "Поделиться",
		SwitchInlineQuery: g.UUID,
	})
	log.Printf("GiveGroupNameHandler END + %+v\n", r)
	return r
}

type OneGroupHandler struct {
	group *domain.Group
}

func (h *OneGroupHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.ClearButtons()
	res := h.group.Name + "\n"
	if len(h.group.Accounts) == 1 {
		res += "В группе кроме вас никого"
	} else {
		for _, acc := range h.group.Accounts {
			res += acc.FirstName + " " + acc.LastName + "\n"
		}
	}
	r.Text = res

	r.AddButton(&bot.Button{
		Text:              "Поделиться",
		SwitchInlineQuery: h.group.UUID,
	})

	r.AddButtonString("Покинуть", &LeaveGroupHandle{h.group})

	if h.group.CreatorAccount.ChatId == c.BotAccount.ChatId {
		r.AddButtonString("Удалить", &DeleteGroupHandler{h.group})
	}

	return r
}

type LeaveGroupHandle struct {
	group *domain.Group
}

func (h *LeaveGroupHandle) Handle(c *bot.Context) *bot.Response {
	err := db.LeaveGroup(h.group, toDomainAccount(c.BotAccount))
	if err != nil {
		panic(err)
	}
	r := c.CurrentResponse
	r.Text = "Вы покинули группу *" + h.group.Name + "*"
	r.ClearButtons()
	return r
}

type DeleteGroupHandler struct {
	group *domain.Group
}

func (h *DeleteGroupHandler) Handle(c *bot.Context) *bot.Response {
	if h.group.CreatorAccount.ChatId != c.BotAccount.ChatId {
		panic("ewq!")
	}
	err := db.DeleteGroup(h.group)
	if err != nil {
		panic(err)
	}
	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Группа *" + h.group.Name + "* удалена"
	return r
}
