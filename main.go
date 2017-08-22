package main

import (
	"smoke3/handlers"
	"smoke3/db"
	"bot/bot"
	"strings"
	"log"
	"os"
	"flag"
	"net/http"
)

func main() {
	runDDL := flag.Bool("runDDL", false, "run ddl?")
	//webhook := flag.String("host", "", "host to enable webhook")
	flag.Parse()
	db.Init(*runDDL)

	bot.Init(os.Getenv("TOKEN"))
	host := os.Getenv("HOST")
	if host != "" {
		if err := bot.EnableWebhook(host); err != nil {
			panic(err)
		}
	}

	bot.RegisterHandler("/start", &handlers.StartHandler{})
	bot.RegisterHandler("/go", &handlers.GoSmokeHandler{})

	bot.RegisterMatchedHandler(&StartMatcher{}, &handlers.StartJoinGroupHandler{})

	bot.RegisterInlineHandler(&ShareGroupInlineHandler{})

	bot.Run()
}

type StartMatcher struct{}

func (m StartMatcher) Match(c *bot.Context) bool {
	return strings.HasPrefix(c.Message.Text, "/start ")
}

type ShareGroupInlineHandler struct{}

func (h *ShareGroupInlineHandler) Handle(c *bot.Context) *bot.InlineAnswer {
	log.Println("ShareGroupInlineHandler START")
	a := &bot.InlineAnswer{
		InlineId: c.Inline.Id,
	}

	g, err := db.GetGroupByUUID(c.Inline.Query)

	if err != nil {
		log.Printf("Couldn't find group by uuid %v\n", c.Inline.Query)
		return nil
	}

	a.Title = g.Name
	a.MessageText = "Я создал группу *" + g.Name + "*. Давай к нам!"
	a.Description = "Нажмите сюда, чтобы поделиться группой"
	a.Button = &bot.Button{
		Text: "Присоедениться",
		URL:  "https://telegram.me/vaping_smoking_bot?start=" + g.UUID,
	}
	log.Println("ShareGroupInlineHandler END")
	return a
}
