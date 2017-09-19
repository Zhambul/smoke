package handlers

import (
	"bot/bot"
	"smoke3/db"
	"smoke3/domain"
	"smoke3/smoke"
	"smoke3/util"
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
	return restoreResponse(c.CurrentResponse, h.Smoke, c.BotAccount.ChatId)
}

type GoSmokeGroupHandler struct {
	group *domain.Group
	min   int
}

func (h *GoSmokeGroupHandler) Handle(c *bot.Context) *bot.Response {
	s := smoke.NewSmoke(h.group, c.BotAccount.ChatId, h.min)

	r := c.CurrentResponse
	r.ClearButtons()
	setCreatorButtons(r, s, c.BotAccount.ChatId)

	s.SCs[c.BotAccount.ChatId].PostResponse = r

	for _, sc := range s.SCs {
		if sc.Account.ChatId == c.BotAccount.ChatId {
			continue
		}
		setRegularButtons(sc.PostResponse, s, sc.Account.ChatId)
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
	return restoreResponse(r, h.Smoke, c.BotAccount.ChatId)
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
		&CancelSmokeHandlerEnd{h.Smoke},
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
	Smoke  *smoke.Smoke
	answer bool
}

func (h *AnswerHandler) Handle(c *bot.Context) *bot.Response {
	h.Smoke.SetAnswer(c.BotAccount, h.answer)
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

type AskForCigaHandler struct {
	Smoke        *smoke.Smoke
	RequesterCtx *smoke.SmokerContext
}

func (h *AskForCigaHandler) Handle(c *bot.Context) *bot.Response {
	go h.Smoke.NotifyOne("Ищем сигаретку", h.RequesterCtx.Account.ChatId, false)
	go askSmokersForCiga(h)
	return nil
}

func askSmokersForCiga(h *AskForCigaHandler) {
	options := make(map[string]bot.Handler)
	options["стрельнуть"] = &AnswerToCigaHandler{Smoke: h.Smoke, RequesterCtx: h.RequesterCtx}
	options["отказать"] = &CancelDialog{h.Smoke}

	for _, sc := range h.Smoke.SCs {
		if sc.Account.ChatId == h.RequesterCtx.Account.ChatId {
			continue
		}
		h.Smoke.LockUserUpdate(util.ToBotAccount(sc.Account))
		msg := h.RequesterCtx.Account.FirstName + " просит стрельнуть сигарету"
		go h.Smoke.AskOne(msg, options, sc)
		go h.Smoke.NotifyOne(msg, sc.Account.ChatId, true)
	}

}

type AnswerToCigaHandler struct {
	Smoke        *smoke.Smoke
	RequesterCtx *smoke.SmokerContext
}

func (h *AnswerToCigaHandler) Handle(c *bot.Context) *bot.Response {
	h.Smoke.UnlockUserUpdate(c.BotAccount)
	go h.Smoke.NotifyOne(h.RequesterCtx.Account.FirstName+" искренне благодарен", c.BotAccount.ChatId, true)
	go h.Smoke.NotifyOne(c.BotAccount.FirstName+" согласился стрельнуть сигарету", h.RequesterCtx.Account.ChatId, false)
	return restoreResponse(c.CurrentResponse, h.Smoke, c.BotAccount.ChatId)
}
