package smoke

import (
	"bot/bot"
	"smoke3/domain"
	"strconv"
	"sync"
	"time"
	"log"
	"smoke3/util"
)

type SmokerContext struct {
	Account      *domain.Account
	PostResponse *bot.Response
	Context      *bot.Context
	Answer       string
	Comment      string
}

type Smoke struct {
	group     *domain.Group
	min       int
	cancelled bool
	SCs       map[int]*SmokerContext
	CreatorSC *SmokerContext
	lock      sync.Mutex
}

func (s *Smoke) SetComment(botAcc *bot.BotAccount, comment string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Printf("setting comment - %v\n", comment)
	s.SCs[botAcc.ChatId].Comment = comment
	s.update()
	go s.notify("*"+botAcc.FirstName+"* - "+comment, botAcc.ChatId)
}

func (s *Smoke) SetAnswer(botAcc *bot.BotAccount, answer string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.SCs[botAcc.ChatId].Answer != answer {
		s.SCs[botAcc.ChatId].Answer = answer
		s.update()
		go s.notify("*"+botAcc.FirstName+"* - "+answer, botAcc.ChatId)
	}
}

func NewSmoke(g *domain.Group, creatorChatId int, min int) *Smoke {
	s := &Smoke{
		min:   min,
		group: g,
		SCs:   make(map[int]*SmokerContext, 0),
	}

	for _, acc := range g.Accounts {
		r := &bot.Response{}
		sc := &SmokerContext{
			Account:      acc,
			Context:      bot.RegisterContext(util.ToBotAccount(acc)),
			PostResponse: r,
		}
		if acc.ChatId == creatorChatId {
			s.CreatorSC = sc
			sc.Answer = "Да"
		} else {
			sc.Answer = "Нет"
		}

		s.SCs[acc.ChatId] = sc
	}
	return s
}

func (s *Smoke) Start() {
	s.lock.Lock()
	s.update()
	s.lock.Unlock()
	go s.timeLoop()
}

func (s *Smoke) timeLoop() {
	t := time.NewTicker(1 * time.Minute)
	for {
		<-t.C
		log.Println("decrementing min")
		s.lock.Lock()
		s.min--
		s.update()
		s.lock.Unlock()
		if s.min == 0 {
			break
		}
		if s.min == 5 {
			go s.notify("Через 5 минут выходить", 0)
		}
	}
}

func (s *Smoke) notify(msg string, omitChatId int) {
	log.Printf("Notify called to %v contexts\n", len(s.SCs))
	s.lock.Lock()
	if s.cancelled {
		return
	}
	s.lock.Unlock()
	for _, smokerContext := range s.SCs {
		if smokerContext.Account.ChatId == omitChatId {
			continue
		}
		r := &bot.Response{
			Text: msg,
		}
		log.Println("Notify!!!!")

		smokerContext.Context.SendReply(r)
		time.Sleep(5 * time.Second)

		smokerContext.Context.DeleteResponse(r)
	}
}

func (s *Smoke) Cancel() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cancelled = true

	for _, smokerContext := range s.SCs {
		smokerContext.Context.DeleteResponse(smokerContext.PostResponse)
	}
}

func (s *Smoke) format() string {
	var when string
	if s.min < 1 {
		when = "сейчас"
	} else {
		when = "через *" + strconv.Itoa(s.min) + "* минут"
	}

	res := "*" + s.CreatorSC.Account.FirstName + "* из группы *" +
		s.group.Name + "*" + " вызывает " + when + "\n\n"

	for _, smokerContext := range s.SCs {
		res += smokerContext.Account.FirstName + " - " + smokerContext.Answer
		if smokerContext.Comment != "" {
			res += ", _" + smokerContext.Comment + "_ "
		}
		res += "\n"
	}

	res += "\n_Ответьте на это сообщение для комментария_"
	return res
}

func (s *Smoke) update() {
	if s.cancelled {
		return
	}

	for _, smokerContext := range s.SCs {
		r := smokerContext.PostResponse
		r.Text = s.format()
		/*
		{"chat_id":104291596,"text":"*Maria* из группы *Тюлени* вызывает через *30* минут\n\nGleb - Нет\nZhambyl - Нет\nMaria - Да\n","message_id":3434,"reply_markup":{"inline_keyboard":[[{"text":"Да","callback_data":"MKcSiStqDUZs"},{"text":"Нет","callback_data":"oLEziJiWmAPK"}],[{"text":"Отменить","callback_data":"vATdjZyEvvms"}]]},"parse_mode":"Markdown"}
panic: HTTP ERROR: url - https://api.telegram.org/bot366621722:AAH5scmfkscK8_Es0dNIJj8gZ-lxluCYD1o/editMessageText
, status code - 400 body - {"ok":false,"error_code":400,"description":"Bad Request: message to edit not found"}

*/
		smokerContext.Context.SendReply(r)
	}
}
