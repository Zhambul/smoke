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
	Going        bool
	Answered     bool
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
	log.Println("Smoke::SetComment START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	s.SCs[botAcc.ChatId].Comment = comment
	log.Println("Smoke::unlock")
	s.lock.Unlock()
	go s.updateWithNotify("*"+botAcc.FirstName+"* - "+comment, botAcc.ChatId)
	log.Println("Smoke::SetComment END")
}

func (s *Smoke) SetAnswer(botAcc *bot.BotAccount, going bool) {
	log.Println("Smoke::SetAnswer START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	s.SCs[botAcc.ChatId].Answered = true
	if s.SCs[botAcc.ChatId].Going != going {
		s.SCs[botAcc.ChatId].Going = going
		go s.updateWithNotify("*"+botAcc.FirstName+"* - "+boolToAnswer(going), botAcc.ChatId)
	}
	log.Println("Smoke::unlock")
	s.lock.Unlock()
	log.Println("Smoke::SetAnswer END")
}

func boolToAnswer(going bool) string {
	if going {
		return "Да"
	}
	return "Нет"
}

func NewSmoke(g *domain.Group, creatorChatId int, min int) *Smoke {
	log.Println("NewSmoke START")
	s := &Smoke{
		min:   min,
		group: g,
		SCs:   make(map[int]*SmokerContext, 0),
	}

	for _, acc := range g.Accounts {
		sc := &SmokerContext{
			Account:      acc,
			Context:      bot.GetContext(util.ToBotAccount(acc)),
			PostResponse: &bot.Response{},
		}
		if acc.ChatId == creatorChatId {
			s.CreatorSC = sc
			s.CreatorSC.Going = true
		}

		s.SCs[acc.ChatId] = sc
	}

	log.Println("NewSmoke END")
	return s
}

func (s *Smoke) Start() {
	go s.update()
	s.timeLoop()
}

func (s *Smoke) timeLoop() {
	log.Println("Smoke::timeLoop START")
	t := time.NewTicker(1 * time.Minute)
	for {
		<-t.C
		log.Println("decrementing min")
		log.Println("Smoke:lock")
		s.lock.Lock()
		s.min--
		log.Println("Smoke:unlock")
		s.lock.Unlock()
		if s.min == 5 {
			if s.goingSmokers() > 1 {
				go s.notifyAll("Группа *"+s.group.Name+"* выходит через 5 минут", 0)
			}
		}
		if s.min <= 0 {
			if s.goingSmokers() > 1 {
				log.Println("Group's going")
				go s.notifyAll("Группа *"+s.group.Name+"* выходит", 0)
			} else {
				log.Println("Group's not going")
				s.Cancel()
			}
			break
		}
		s.update()

	}
	log.Println("Smoke::timeLoop END")
}

func (s *Smoke) notifyOne(msg string, smokerContext *SmokerContext) {
	log.Println("Smoke::notifyOne START")
	if s.cancelled {
		return
	}

	if !s.SCs[smokerContext.Account.ChatId].Going {
		return
	}

	r := &bot.Response{
		Text: msg,
	}

	smokerContext.Context.SendReply(r)
	time.Sleep(5 * time.Second)
	smokerContext.Context.DeleteResponse(r)
	log.Println("Smoke::notifyOne END")
}

func (s *Smoke) goingSmokers() int {
	log.Println("Smoke::goingSmokers START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	goingSmokers := 0
	for _, sc := range s.SCs {
		if sc.Going {
			goingSmokers++
		}
	}
	log.Println("Smoke::unlock")
	s.lock.Unlock()
	log.Println("Smoke::goingSmokers END")
	return goingSmokers
}

func (s *Smoke) notifyAll(msg string, omitChatId int) {
	for _, smokerContext := range s.SCs {
		if smokerContext.Account.ChatId == omitChatId || !smokerContext.Going {
			continue
		}
		go s.notifyOne(msg, smokerContext)
	}
}

func (s *Smoke) Cancel() {
	log.Println("Smoke::Cancel START")
	log.Println("Smoke::lock")

	s.lock.Lock()
	if s.cancelled {
		return
	}
	s.cancelled = true

	for _, smokerContext := range s.SCs {
		go smokerContext.Context.DeleteResponse(smokerContext.PostResponse)
	}
	log.Println("Smoke::unlock")
	s.lock.Unlock()
	log.Println("Smoke::Cancel END")
}

func (s *Smoke) format() string {
	log.Println("Smoke::format START")
	var when string
	if s.min < 1 {
		when = "сейчас"
	} else {
		when = "через *" + strconv.Itoa(s.min) + "* минут"
	}

	res := "*" + s.CreatorSC.Account.FirstName + "* из группы *" +
		s.group.Name + "*" + " вызывает " + when + "\n\n"

	for _, sc := range s.SCs {
		if sc.Answered {
			res += sc.Account.FirstName + " - " + boolToAnswer(sc.Going)
		} else {
			res += sc.Account.FirstName + " - "
		}

		if sc.Comment != "" {
			res += ", _" + sc.Comment + "_ "
		}
		res += "\n"
	}

	res += "\n_Ответьте на это сообщение для комментария_"
	log.Println("Smoke::format END")
	return res
}

func (s *Smoke) update() {
	log.Println("Smoke::update START")
	s.updateWithNotify("", 0)
	log.Println("Smoke::update END")
}

func (s *Smoke) updateWithNotify(msg string, omitChatId int) {
	log.Println("Smoke::updateWithNotify START")
	if s.cancelled {
		return
	}

	for _, smokerContext := range s.SCs {
		r := smokerContext.PostResponse
		r.Text = s.format()
		go smokerContext.Context.SendReply(r)
		if msg != "" {
			if smokerContext.Account.ChatId != omitChatId {
				go s.notifyOne(msg, smokerContext)
			}
		}
	}
	log.Println("Smoke::updateWithNotify END")
}
