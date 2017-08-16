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
			s.CreatorSC.Answered = true
		}

		s.SCs[acc.ChatId] = sc
	}

	log.Println("NewSmoke END")
	return s
}

func (s *Smoke) Start() {
	go s.update()
	go s.lifecycle()
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
		go smokerContext.Context.Send(r)
		if msg != "" {
			if smokerContext.Account.ChatId != omitChatId {
				go s.notifyOne(msg, smokerContext)
			}
		}
	}
	log.Println("Smoke::updateWithNotify END")
}
