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
	s.SCs[botAcc.ChatId].Comment = comment
	s.lock.Unlock()
	go s.updateWithNotify("*"+botAcc.FirstName+"* - "+comment, botAcc.ChatId)
}

func (s *Smoke) SetAnswer(botAcc *bot.BotAccount, answer string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.SCs[botAcc.ChatId].Answer != answer {
		s.SCs[botAcc.ChatId].Answer = answer
		go s.updateWithNotify("*"+botAcc.FirstName+"* - "+answer, botAcc.ChatId)
	}
}

func NewSmoke(g *domain.Group, creatorChatId int, min int) *Smoke {
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
		}

		s.SCs[acc.ChatId] = sc
	}

	return s
}

func (s *Smoke) Start() {
	go s.update()
	go s.timeLoop()
}

func (s *Smoke) timeLoop() {
	t := time.NewTicker(1 * time.Minute)
	for {
		<-t.C
		log.Println("decrementing min")
		s.lock.Lock()
		s.min--
		go s.update()
		s.lock.Unlock()
		if s.min == 0 {
			go s.notifyAll("Группа *"+s.group.Name+"* выходит", 0)
			break
		}
		if s.min == 5 {
			go s.notifyAll("Группа *"+s.group.Name+"* выходит через 5 минут", 0)
		}
	}
}

func (s *Smoke) notifyOne(msg string, smokerContext *SmokerContext) {
	if s.cancelled {
		return
	}

	r := &bot.Response{
		Text: msg,
	}

	smokerContext.Context.SendReply(r)
	time.Sleep(5 * time.Second)
	smokerContext.Context.DeleteResponse(r)
}

func (s *Smoke) notifyAll(msg string, omitChatId int) {
	for _, smokerContext := range s.SCs {
		if smokerContext.Account.ChatId == omitChatId {
			continue
		}
		go s.notifyOne(msg, smokerContext)
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
	s.updateWithNotify("", 0)
}

func (s *Smoke) updateWithNotify(msg string, chatId int) {
	if s.cancelled {
		return
	}

	for _, smokerContext := range s.SCs {
		r := smokerContext.PostResponse
		r.Text = s.format()
		go smokerContext.Context.SendReply(r)
		if msg != "" {
			if smokerContext.Account.ChatId != chatId {
				go s.notifyOne(msg, smokerContext)
			}
		}
	}
}
