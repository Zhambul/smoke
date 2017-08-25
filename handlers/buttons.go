package handlers

import (
	"bot/bot"
	"strconv"
	"smoke3/smoke"
)

func setTimeButtons(r *bot.Response, getHandler func(min int) bot.Handler, cancelHandler bot.Handler) {
	r.AddButton(&bot.Button{
		Text:    "Сейчас",
		Handler: getHandler(0),
	})
	r.AddButtonRow(timeButton(5, getHandler), timeButton(10, getHandler), timeButton(15, getHandler))
	r.AddButtonRow(timeButton(20, getHandler), timeButton(30, getHandler), timeButton(40, getHandler))
	r.AddButtonString("Отменить", cancelHandler)
}

func timeButton(min int, getHandler func(min int) bot.Handler) *bot.Button {
	return &bot.Button{
		Text:    strconv.Itoa(min),
		Handler: getHandler(min),
	}
}

func setYesNoButtons(r *bot.Response, yesHandler bot.Handler, noHandler bot.Handler) {
	r.AddButtonRow(&bot.Button{
		Text:"Да",
		Handler:yesHandler,
	},
		&bot.Button{
		Text:"Нет",
		Handler:noHandler,
	})
}

func setCreatorButtons(r *bot.Response, s *smoke.Smoke) {
	r.AddButtonString("Изменить время", &ChangeTimeHandlerStart{s})
	r.AddButtonString("Отменить", &CancelSmokeHandlerStart{s})
	r.ReplyHandler = &ReplyHandler{s}
}

func setRegularButtons(r *bot.Response, s *smoke.Smoke) {
	a := &AnswerHandler{
		Smoke: s,
	}
	setYesNoButtons(r, a, a)
	r.AddButtonString("Предложить другое время", &SuggestTimeHandlerStart{s})
	r.ReplyHandler = &ReplyHandler{s}
}

func restoreRegularResponse(r *bot.Response, s *smoke.Smoke) *bot.Response{
	r.Text = s.Format()
	r.ClearButtons()
	setRegularButtons(r, s)
	return r
}

func restoreCreatorResponse(r *bot.Response, s *smoke.Smoke) *bot.Response {
	r.Text = s.Format()
	r.ClearButtons()
	setCreatorButtons(r, s)
	return r
}