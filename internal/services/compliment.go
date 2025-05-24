package services

import (
	"bufio"
	"log"
	"math/rand/v2"
	"os"
	"strings"
)

type ComplimentService struct {
	compliments []string
}

func NewComplimentService(path string) *ComplimentService {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("services: cannot open compliments file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if txt == "" {
			continue
		}
		lines = append(lines, "♥"+txt+"♥")
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("services: reading file: %w", err)
	}
	if len(lines) == 0 {
		log.Fatal("services: no compliments found in file")
	}

	return &ComplimentService{compliments: lines}
}

func (s *ComplimentService) Random() string {
	idx := rand.IntN(len(s.compliments))
	return s.compliments[idx]
}
