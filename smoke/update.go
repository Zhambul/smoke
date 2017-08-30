package smoke

import (
	"bot/bot"
	"log"
	"sort"
	"strconv"
	"time"
)

func (s *Smoke) update() {
	log.Println("Smoke::update START")
	s.updateWithNotify("", 0)
	log.Println("Smoke::update END")
}

func (s *Smoke) updateWithNotify(msg string, omitChatId int) {
	log.Println("Smoke::updateWithNotify START")

	for _, sc := range s.SCs {
		if sc.Locked {
			continue
		}

		r := sc.PostResponse
		r.Text = s.Format()
		go sc.Context.Send(r)
		if msg != "" {
			if sc.Account.ChatId != omitChatId {
				go s.NotifyOne(msg, sc, true)
			}
		}
	}
	log.Println("Smoke::updateWithNotify END")
}

func (s *Smoke) Format() string {
	log.Println("Smoke::Format START")
	var when string
	if s.min < 1 {
		when = "*сейчас*"
	} else {
		when = "через *" + strconv.Itoa(s.min) + "* минут"
	}

	res := "*" + s.GetUniqueUserName(s.CreatorSC.Account) + "* из группы *" +
		s.group.Name + "*" + " вызывает " + when + "\n\n"

	var keys []int
	for chatId := range s.SCs {
		keys = append(keys, chatId)
	}

	sort.Ints(keys)

	for _, chatId := range keys {
		sc := s.SCs[chatId]
		res += s.answer(sc)
		res += s.comment(sc)
		res += "\n"
	}

	res += "\n_Ответьте на это сообщение для комментария_"
	log.Println("Smoke::Format END")
	return res
}

func (s *Smoke) answer(sc *SmokerContext) string {
	if sc.Answered {
		return s.GetUniqueUserName(sc.Account) + " - " + boolToAnswer(sc.Going)
	}
	return s.GetUniqueUserName(sc.Account) + " - "
}

func (s *Smoke) comment(sc *SmokerContext) string {
	if sc.Comment != "" {
		return ", _" + sc.Comment + "_ "
	}
	return ""
}

func (s *Smoke) NotifyOne(msg string, sc *SmokerContext, onlyGoing bool) {
	log.Println("Smoke::NotifyOne START")

	if onlyGoing && !s.SCs[sc.Account.ChatId].Going {
		return
	}

	log.Printf("Smoke::notifying msg - %v, %v\n", msg, sc)

	r := &bot.Response{
		Text: msg,
	}

	sc.Context.Send(r)
	time.Sleep(15 * time.Second)
	sc.Context.DeleteResponse(r)
	log.Println("Smoke::NotifyOne END")
}

func (s *Smoke) AskOne(msg string, resposeOptions map[string]bot.Handler, sc *SmokerContext) {
	log.Println("Smoke::AskOne START")

	if !s.SCs[sc.Account.ChatId].Going {
		return
	}

	log.Printf("Smoke::ask msg - %v, %v\n", msg, sc)

	r := &bot.Response{
		Text: msg,
	}

	for label, handler := range resposeOptions {
		r.AddButton(&bot.Button{
			Text:    label,
			Handler: handler,
		})
	}

	sc.Context.Send(r)
	log.Println("Smoke::AskOne END")
}

func (s *Smoke) goingSmokers() int {
	log.Println("Smoke::goingSmokers START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	goingSmokers := 0
	for _, sc := range s.SCs {
		if sc.Going {
			log.Printf("Smoke::goingSmokers. %v is going\n", sc)
			goingSmokers++
		}
	}
	log.Println("Smoke::unlock")
	s.lock.Unlock()
	log.Printf("Smoke::goingSmokers END. %v\n", goingSmokers)
	return goingSmokers
}

func boolToAnswer(going bool) string {
	if going {
		return "Да"
	}
	return "Нет"
}

func (s *Smoke) decrement() {
	log.Println("Smoke:lock")
	s.lock.Lock()
	log.Println("decrementing min")
	s.min--
	log.Println("Smoke:unlock")
	s.lock.Unlock()
}

func (s *Smoke) notifyAll(msg string) {
	s.notifyAllExcept(msg, 0)
}

func (s *Smoke) notifyAllExcept(msg string, omitChatId int) {
	for _, smokerContext := range s.SCs {
		if smokerContext.Account.ChatId == omitChatId || !smokerContext.Going {
			continue
		}
		go s.NotifyOne(msg, smokerContext, true)
	}
}

func (s *Smoke) AskAllExcept(msg string, responseOptions map[string]bot.Handler, omitChatId int) {
	for _, smokerContext := range s.SCs {
		if smokerContext.Account.ChatId == omitChatId || !smokerContext.Going || smokerContext.Locked {
			continue
		}
		go s.AskOne(msg, responseOptions, smokerContext)
	}
}

func (s *Smoke) delayedCancel(min int) {
	log.Println("Smoke::delayedCancel START")
	s.delayedCancelEnabled = true
	defer func() {
		s.delayedCancelEnabled = false
	}()
	t := time.NewTicker(time.Duration(min) * time.Minute)
	select {
	case <-t.C:
		log.Println("Smoke::delayedCancel END")
		s.Cancel(false)
	case <-s.cancelDelayedCancel:
		log.Println("Smoke::delayedCancel END. cancelLifecycle")
		return
	}
}
