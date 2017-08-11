package handlers

import (
	"smoke3/db"
	"smoke3/domain"
	"bot/bot"
	"strconv"
	"smoke3/smoke"
)

type GoSmokeHandler struct {
}

func (t *GoSmokeHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.ClearButtons()

	groups, err := db.GetGroupsByAccount(toDomainAccount(c.BotAccount))

	if err != nil {
		panic(err)
	}

	if len(groups) == 0 {
		r.Text = "У вас нет групп"
		return r
	}

	r.Text = "Выберите группу"

	for _, g := range groups {
		r.AddButtonString(g.Name, &ChooseTimeHandler{g})
	}
	r.AddButtonString("Назад", &StartHandler{})

	return r
}

type ChooseTimeHandler struct {
	group *domain.Group
}

func (h *ChooseTimeHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.Text = "Через сколько минут?"
	r.ClearButtons()
	r.AddButtonRow(h.goSmokeButton(5), h.goSmokeButton(10), h.goSmokeButton(15))
	r.AddButtonRow(h.goSmokeButton(20), h.goSmokeButton(30), h.goSmokeButton(40))
	return r
}

func (h *ChooseTimeHandler) goSmokeButton(min int) *bot.Button {
	return &bot.Button{
		Text: strconv.Itoa(min),
		Handler: &GoSmokeGroupHandler{
			min:   min,
			group: h.group,
		},
	}
}

type GoSmokeGroupHandler struct {
	group *domain.Group
	min   int
}

func (h *GoSmokeGroupHandler) Handle(c *bot.Context) *bot.Response {
	s := smoke.NewSmoke(h.group, c.BotAccount.ChatId, h.min)
	c.CurrentResponse.ClearButtons()
	s.SCs[c.BotAccount.ChatId].PostResponse = c.CurrentResponse

	for _, smokerContext := range s.SCs {
		a := &AnswerHandler{
			Smoke: s,
		}

		smokerResponse := smokerContext.PostResponse
		smokerResponse.AddButtonRow(&bot.Button{Handler: a, Text: "Да"}, &bot.Button{Handler: a, Text: "Нет"})

		if smokerContext.Account.ChatId == c.BotAccount.ChatId {
			smokerResponse.AddButton(&bot.Button{Text: "Отменить", Handler: &CancelSmokeHandler{
				Smoke: s,
			}})
		}
		smokerResponse.ReplyHandler = &ReplyHandler{
			Smoke: s,
		}
	}
	s.Start()

	return nil
}

type ReplyHandler struct {
	Smoke *smoke.Smoke
}

func (h *ReplyHandler) Handle(c *bot.Context) *bot.Response {
	h.Smoke.SetComment(c.BotAccount, c.Message.Text)
	return nil
}

type CancelSmokeHandler struct {
	Smoke *smoke.Smoke
}

func (h *CancelSmokeHandler) Handle(c *bot.Context) *bot.Response {
	h.Smoke.Cancel()
	return nil
}

type AnswerHandler struct {
	Yes   *bot.Button
	No    *bot.Button
	Smoke *smoke.Smoke
}

func (h *AnswerHandler) Handle(c *bot.Context) *bot.Response {
	h.Smoke.SetAnswer(c.BotAccount, c.CurrentResponse.ClickedButton.Text)
	return nil
}

func toDomainAccount(botAcc *bot.BotAccount) *domain.Account {
	acc, err := db.GetAccountByChatId(botAcc.ChatId)
	if err != nil {
		if err == db.NotFound {
			acc, err := db.CreateAccount(botAcc.FirstName, botAcc.LastName, botAcc.ChatId)
			if err != nil {
				panic(err)
			}
			return acc
		}
		panic(err)
	}
	return acc
}
