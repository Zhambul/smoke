package handlers

import (
	"smoke3/db"
	"smoke3/domain"
	"bot/bot"
	"strconv"
	"smoke3/smoke"
	"smoke3/util"
	"strings"
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
	if len(h.group.Accounts) < 2 {
		r := c.CurrentResponse
		r.Text = "В группе *" + h.group.Name + "* кроме вас никого"
		r.ClearButtons()
		r.AddButton(util.ShareButton(h.group))
		r.AddButtonString("Удалить", &DeleteGroupHandler{group: h.group})
		return r
	}

	r := c.CurrentResponse
	r.Text = "Через сколько минут?"
	r.ClearButtons()
	r.AddButton(&bot.Button{
		Text: "Сейчас",
		Handler: &GoSmokeGroupHandler{
			min:   0,
			group: h.group,
		},
	})
	r.AddButtonRow(h.goSmokeButton(5), h.goSmokeButton(10), h.goSmokeButton(15))
	r.AddButtonRow(h.goSmokeButton(20), h.goSmokeButton(30), h.goSmokeButton(40))
	r.AddButtonString("Отменить", &StartHandler{})
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

		sr := smokerContext.PostResponse
		sr.AddButtonRow(&bot.Button{Handler: a, Text: "Да"}, &bot.Button{Handler: a, Text: "Нет"})

		if smokerContext.Account.ChatId == c.BotAccount.ChatId {
			setCreatorButtons(sr, s)
		}
		sr.ReplyHandler = &ReplyHandler{
			Smoke: s,
		}
	}
	go s.Start()
	return nil
}

func setCreatorButtons(sr *bot.Response, s *smoke.Smoke)  {
	sr.AddButtonString("Отменить", &CancelSmokeHandler{
		Smoke: s,
	})
	sr.AddButtonString("Изменить время", &ChangeTimeHandlerStart{
		Smoke: s,
	})
}

type ReplyHandler struct {
	Smoke *smoke.Smoke
}

func (h *ReplyHandler) Handle(c *bot.Context) *bot.Response {
	h.Smoke.SetComment(c.BotAccount, c.Message.Text)
	return nil
}

type ChangeTimeHandlerStart struct {
	Smoke *smoke.Smoke
}

func (h *ChangeTimeHandlerStart) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.Text = "Через сколько минут?"
	r.ClearButtons()
	r.AddButton(&bot.Button{
		Text: "Сейчас",
		Handler: &ChangeTimeHandlerEnd{
			min:   0,
			Smoke: h.Smoke,
		},
	})
	r.AddButtonRow(h.changeTimeButton(5), h.changeTimeButton(10), h.changeTimeButton(15))
	r.AddButtonRow(h.changeTimeButton(20), h.changeTimeButton(30), h.changeTimeButton(40))
	r.AddButtonString("Отменить", &StartHandler{})
	return r
}

type ChangeTimeHandlerEnd struct {
	Smoke *smoke.Smoke
	min   int
}

func (h *ChangeTimeHandlerEnd) Handle(c *bot.Context) *bot.Response {
	h.Smoke.ChangeTime(h.min)
	r := h.Smoke.CreatorSC.PostResponse
	r.ClearButtons()
	setCreatorButtons(r, h.Smoke)
	return nil
}

func (h *ChangeTimeHandlerStart) changeTimeButton(min int) *bot.Button {
	return &bot.Button{
		Text: strconv.Itoa(min),
		Handler: &ChangeTimeHandlerEnd{
			min:   min,
			Smoke: h.Smoke,
		},
	}
}

type CancelSmokeHandler struct {
	Smoke *smoke.Smoke
}

func (h *CancelSmokeHandler) Handle(c *bot.Context) *bot.Response {
	go h.Smoke.Cancel()
	return nil
}

type AnswerHandler struct {
	Yes   *bot.Button
	No    *bot.Button
	Smoke *smoke.Smoke
}

func (h *AnswerHandler) Handle(c *bot.Context) *bot.Response {
	answer := c.CurrentResponse.ClickedButton.Text
	answer = strings.TrimSpace(answer)
	h.Smoke.SetAnswer(c.BotAccount, answer == "Да")
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
