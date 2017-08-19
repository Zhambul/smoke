package handlers

import "bot/bot"

type StartHandler struct {
}

func (t *StartHandler) Handle(c *bot.Context) *bot.Response {
	r := c.CurrentResponse
	r.ClearButtons()
	r.Text = "Слушаю"

	r.AddButtonString("Го", &GoSmokeHandler{})
	r.AddButtonString("Меню", &MenuHandler{})
	return r
}