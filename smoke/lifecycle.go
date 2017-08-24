package smoke

import (
	"time"
	"log"
)

func (s *Smoke) lifecycle() {
	log.Println("Smoke::lifecycle START")
	s.lifecycleEnabled = true
	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Smoke::lifecycle. Tick")
			if end := s.tick(); end {
				s.lifecycleEnabled = false
				log.Println("Smoke::lifecycle END")
				return
			}
		case <-s.cancelLifecycle:
			s.lifecycleEnabled = false
			log.Println("Smoke::lifecycle END. Cancel")
			return
		}
	}
	s.lifecycleEnabled = false
	log.Println("Smoke::lifecycle END")
}

func (s *Smoke) tick() (end bool) {
	s.decrement()

	if s.min == 5 {
		s.onFiveMinutes()
	}

	if s.min <= 0 {
		s.onZeroMinutes()
		end = true
	}

	s.update()
	return
}

func (s *Smoke) onFiveMinutes() {
	log.Println("On five minutes")
	if s.goingSmokers() > 1 {
		go s.notifyAll("Группа *" + s.group.Name + "* выходит через 5 минут")
	}
}

func (s *Smoke) onZeroMinutes() {
	log.Println("On zero minutes")
	if s.goingSmokers() > 1 {
		log.Println("Group's going")
		go s.notifyAll("Группа *" + s.group.Name + "* выходит")
	}
	go s.delayedCancel(10)
}
