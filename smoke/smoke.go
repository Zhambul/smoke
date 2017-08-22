package smoke

import (
	"bot/bot"
	"smoke3/domain"
	"sync"
	"log"
	"smoke3/util"
	"strconv"
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
	group           *domain.Group
	min             int
	cancelled       bool
	cancelLifecycle chan bool
	SCs             map[int]*SmokerContext
	CreatorSC       *SmokerContext
	lock            sync.Mutex
}

func NewSmoke(g *domain.Group, creatorChatId int, min int) *Smoke {
	log.Println("NewSmoke START")
	s := &Smoke{
		min:             min,
		group:           g,
		SCs:             make(map[int]*SmokerContext, 0),
		cancelLifecycle: make(chan bool),
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

func (s *Smoke) getUniqueUserName(acc *domain.Account) string {
	for _, sc := range s.SCs {
		if sc.Account.ChatId != acc.ChatId {
			if sc.Account.FirstName == acc.FirstName {
				return acc.FirstName + " " + acc.LastName
			}
		}
	}
	return acc.FirstName
}

func (s *Smoke) ChangeTime(min int) {
	log.Println("Smoke::ChangeTime START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	defer func() {
		log.Println("Smoke::unlock")
		s.lock.Unlock()
		log.Println("Smoke::ChangeTime END")
	}()
	s.cancelLifecycle <- true
	s.min = min
	s.cancelLifecycle = make(chan bool)
	go s.lifecycle()
	go s.updateWithNotify("*" + s.CreatorSC.Account.FirstName+
		"* изменил время на *"+ strconv.Itoa(min)+ "* минут", s.CreatorSC.Account.ChatId)
}

func (s *Smoke) Cancel() {
	log.Println("Smoke::Cancel START")
	log.Println("Smoke::lock")
	s.lock.Lock()
	defer func() {
		log.Println("Smoke::unlock")
		s.lock.Unlock()
		log.Println("Smoke::Cancel END")
	}()

	if s.cancelled {
		return
	}
	s.cancelled = true

	for _, sc := range s.SCs {
		go sc.Context.DeleteResponse(sc.PostResponse)
	}
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
	defer func() {
		log.Println("Smoke::unlock")
		s.lock.Unlock()
		log.Println("Smoke::SetAnswer END")
	}()

	if s.SCs[botAcc.ChatId].Going != going || !s.SCs[botAcc.ChatId].Answered {
		s.SCs[botAcc.ChatId].Answered = true
		log.Println("Smoke::SetAnswer going changed")
		s.SCs[botAcc.ChatId].Going = going
		go s.updateWithNotify("*"+botAcc.FirstName+"* - "+boolToAnswer(going), botAcc.ChatId)
	} else {
		log.Println("Smoke::SetAnswer going not changed")
	}
}
