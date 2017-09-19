package handlers

import (
	"bot/bot"
	"fmt"
	"smoke3/smoke"
	"smoke3/util"
	"strconv"
)

type SuggestTimeHandlerStart struct {
	Smoke *smoke.Smoke
}

func (h *SuggestTimeHandlerStart) Handle(c *bot.Context) *bot.Response {
	h.Smoke.LockUserUpdate(c.BotAccount)
	r := c.CurrentResponse
	r.Text = "Через сколько минут?"
	r.ClearButtons()
	setTimeButtons(r, func(min int) bot.Handler {
		return &SuggestTimeHandlerEnd{
			min:   min,
			Smoke: h.Smoke,
		}
	}, &CancelDialog{h.Smoke})
	return r
}

type SuggestTimeHandlerEnd struct {
	Smoke *smoke.Smoke
	min   int
}

func (h *SuggestTimeHandlerEnd) Handle(c *bot.Context) *bot.Response {
	h.Smoke.UnlockUserUpdate(c.BotAccount)
	go h.Smoke.NotifyOne("Предложение изменить на "+strconv.Itoa(h.min)+" минут отправлено", c.BotAccount.ChatId, true)
	go askCreator(h, c)
	return restoreResponse(c.CurrentResponse, h.Smoke, c.BotAccount.ChatId)
}

func askCreator(h *SuggestTimeHandlerEnd, c *bot.Context) {
	h.Smoke.LockUserUpdate(util.ToBotAccount(h.Smoke.CreatorSC.Account))
	cc := h.Smoke.CreatorSC.Context

	cr := h.Smoke.CreatorSC.PostResponse
	cr.Text = fmt.Sprintf("*%v* предлагает через *%v* минут", h.Smoke.GetUniqueUserName(toDomainAccount(c.BotAccount)), h.min)
	cr.ClearButtons()

	setYesNoButtons(cr, &SuggestTimeHandlerApproved{
		Smoke:     h.Smoke,
		min:       h.min,
		Suggester: c.BotAccount,
	}, &CancelDialog{h.Smoke})

	go h.Smoke.NotifyOne("Предложение изменить время", h.Smoke.CreatorSC.Account.ChatId, true)

	cc.Send(cr)
}

type SuggestTimeHandlerApproved struct {
	Smoke     *smoke.Smoke
	min       int
	Suggester *bot.BotAccount
}

func (h *SuggestTimeHandlerApproved) Handle(c *bot.Context) *bot.Response {
	h.Smoke.UnlockUserUpdate(c.BotAccount)
	//suggester - yes
	h.Smoke.SetAnswer(h.Suggester, true)
	//change time
	h.Smoke.ChangeTime(h.min, h.Suggester)
	return restoreResponse(c.CurrentResponse, h.Smoke, c.BotAccount.ChatId)
}
