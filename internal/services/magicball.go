package services

import (
	"bytes"
	"io"
	"math/rand/v2"
	"os"
)

type MagicBallService struct {
	answers    [20]string
	imagesPath string
}

func NewMagicBallService(imagesPath string) *MagicBallService {
	return &MagicBallService{
		answers: [20]string{
			"As I see it, yes",
			"Ask again later",
			"Better not tell you now",
			"Cannot predict now",
			"Concentrate and ask again",
			"Don’t count on it",
			"It is certain",
			"It is decidedly so",
			"Most likely",
			"My reply is no",
			"My sources say no",
			"Outlook good",
			"Outlook not so good",
			"Reply hazy, try again",
			"Signs point to yes",
			"Very doubtful",
			"Without a doubt",
			"Yes — definitely",
			"Yes",
			"You may rely on it",
		},
		imagesPath: imagesPath,
	}
}

func (s *MagicBallService) answer() string {
	idx := rand.IntN(len(s.answers))
	return s.answers[idx]
}

func (s *MagicBallService) ImageAnswer() (io.Reader, error) {
	fileName := s.imagesPath + s.answer() + ".webp"

	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(file), nil
}
