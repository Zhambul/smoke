package smoke

import (
	"testing"
	"smoke3/domain"
	"bot/bot"
	"time"
)

func Test_smoke(t *testing.T) {
	creator := &domain.Account{
		Id:     12,
		ChatId: 12,
	}

	g := &domain.Group{
		CreatorAccount: creator,
		Accounts:       make([]*domain.Account, 0),
	}

	g.Accounts = append(g.Accounts, creator)
	g.Accounts = append(g.Accounts, &domain.Account{
		Id:     13,
		ChatId: 13,
	})

	g.Accounts = append(g.Accounts, &domain.Account{
		Id:     14,
		ChatId: 14,
	})

	s := NewSmoke(g, creator.ChatId, 0)
	s.SetAnswer(&bot.BotAccount{
		ChatId: 12,
	}, true)

	s.SetAnswer(&bot.BotAccount{
		ChatId: 13,
	}, true)
	s.Start()
}

func Test_first_no(t *testing.T) {
	creator := &domain.Account{
		Id:     12,
		ChatId: 12,
	}

	g := &domain.Group{
		CreatorAccount: creator,
		Accounts:       make([]*domain.Account, 0),
	}

	g.Accounts = append(g.Accounts, creator)
	gleb := &domain.Account{
		Id:     13,
		ChatId: 13,
	}
	g.Accounts = append(g.Accounts, gleb)

	s := NewSmoke(g, creator.ChatId, 0)
	s.Start()
	s.SetAnswer(&bot.BotAccount{
		ChatId: 13,
	}, false)
	time.Sleep(1 * time.Second)

	s.SetAnswer(&bot.BotAccount{
		ChatId: 13,
	}, true)
	//
	//s.SetComment(&bot.BotAccount{
	//		ChatId: 13,
	//	}, "qwe")
	time.Sleep(1 * time.Hour)
}
