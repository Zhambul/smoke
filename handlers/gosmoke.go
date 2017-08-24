package handlers

import (
	"smoke3/db"
	"smoke3/domain"
	"bot/bot"
	"strconv"
	"smoke3/smoke"
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
	//if len(h.group.Accounts) < 2 {
	//	r := c.CurrentResponse
	//	r.Text = "В группе *" + h.group.Name + "* кроме вас никого"
	//	r.ClearButtons()
	//	r.AddButton(util.ShareButton(h.group))
	//	r.AddButtonString("Удалить", &DeleteGroupHandler{group: h.group})
	//	return r
	//}
	r := c.CurrentResponse
	r.Text = "Через сколько минут?"
	r.ClearButtons()
	setTimeButtons(r, func(min int) bot.Handler {
		return &GoSmokeGroupHandler{
			min:   min,
			group: h.group,
		}
	}, &StartHandler{})
	return r
}

type CancelDialog struct {
	Smoke *smoke.Smoke
}

func (h *CancelDialog) Handle(c *bot.Context) *bot.Response {
	h.Smoke.UnlockUserUpdate(c.BotAccount)
	return restoreCreatorResponse(c.CurrentResponse, h.Smoke)
}

type GoSmokeGroupHandler struct {
	group *domain.Group
	min   int
}

func (h *GoSmokeGroupHandler) Handle(c *bot.Context) *bot.Response {
	s := smoke.NewSmoke(h.group, c.BotAccount.ChatId, h.min)

	r := c.CurrentResponse
	r.ClearButtons()
	setCreatorButtons(r, s)

	s.SCs[c.BotAccount.ChatId].PostResponse = r

	for _, sc := range s.SCs {
		if sc.Account.ChatId == c.BotAccount.ChatId {
			continue
		}

		setRegularButtons(sc.PostResponse, s)
	}

	go s.Start()
	return nil
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
	h.Smoke.LockUserUpdate(c.BotAccount)
	r := c.CurrentResponse
	r.Text = "Через сколько минут?"
	r.ClearButtons()
	setTimeButtons(r, func(min int) bot.Handler {
		return &ChangeTimeHandlerEnd{
			min:   min,
			Smoke: h.Smoke,
		}
	}, &CancelDialog{h.Smoke})
	return r
}

type ChangeTimeHandlerEnd struct {
	Smoke *smoke.Smoke
	min   int
}

func (h *ChangeTimeHandlerEnd) Handle(c *bot.Context) *bot.Response {
	h.Smoke.UnlockUserUpdate(c.BotAccount)
	h.Smoke.ChangeTime(h.min, c.BotAccount)
	r := h.Smoke.CreatorSC.PostResponse
	return restoreCreatorResponse(r, h.Smoke)
}

type CancelSmokeHandlerStart struct {
	Smoke *smoke.Smoke
}

func (h *CancelSmokeHandlerStart) Handle(c *bot.Context) *bot.Response {
	h.Smoke.LockUserUpdate(c.BotAccount)
	r := c.CurrentResponse
	r.Text = "Вы уверены, что хотите отменить?"
	r.ClearButtons()
	setYesNoButtons(r,
		&CancelSmokeHandlerEnd{h.Smoke,},
		&CancelDialog{h.Smoke})
	return r
}

type CancelSmokeHandlerEnd struct {
	Smoke *smoke.Smoke
}

func (h *CancelSmokeHandlerEnd) Handle(c *bot.Context) *bot.Response {
	h.Smoke.UnlockUserUpdate(c.BotAccount)
	go h.Smoke.Cancel(true)
	return nil
}

type CancelSmokeHandlerCancel struct {
	Smoke *smoke.Smoke
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
