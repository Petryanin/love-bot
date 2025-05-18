package services

import (
	"math/rand/v2"
)

type MagicBallService struct {
	answers [20]string
}

func NewMagicBallService() *MagicBallService {
	return &MagicBallService{
		answers: [20]string{
			"Бесспорно",
			"Предрешено",
			"Никаких сомнений",
			"Определённо да",
			"Можешь быть уверен в этом",

			"Мне кажется — да",
			"Вероятнее всего",
			"Хорошие перспективы",
			"Знаки говорят — да",
			"Да",

			"Пока не ясно, попробуй снова",
			"Спроси позже",
			"Лучше не рассказывать сейчас",
			"Сейчас нельзя предсказать",
			"Сконцентрируйся и спроси опять",

			"Даже не думай",
			"Мой ответ — «нет»",
			"По моим данным — «нет»",
			"Перспективы не очень хорошие",
			"Весьма сомнительно",
		},
	}
}

func (s *MagicBallService) Answer() string {
	idx := rand.IntN(len(s.answers))
	return s.answers[idx]
}
