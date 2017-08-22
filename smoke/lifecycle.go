package smoke

import (
	"time"
	"log"
)

func (s *Smoke) lifecycle() {
	log.Println("Smoke::lifecycle START")
	if s.min == 0 {
		log.Println("Smoke::lifecycle END")
		go s.delayedCancel(10)
		return
	}
	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Smoke::lifecycle. Tick")
			if end := s.tick(); end {
				break
			}
		case <-s.cancelLifecycle:
			log.Println("Smoke::lifecycle END. Cancel")
			return
		}
	}
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
	if s.goingSmokers() > 1 {
		go s.notifyAll("Группа *" + s.group.Name + "* выходит через 5 минут")
	}
}

func (s *Smoke) onZeroMinutes() {
	if s.goingSmokers() > 1 {
		log.Println("Group's going")
		go s.notifyAll("Группа *" + s.group.Name + "* выходит")
		go s.delayedCancel(10)
	} else {
		log.Println("Group's not going")
		go s.Cancel()
	}
}
