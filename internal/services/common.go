package services

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Pluralize подбирает форму слова в зависимости от числа (для русских окончаний)
func Pluralize(n int, one, few, many string) string {
	nAbs := n % 100
	if nAbs >= 11 && nAbs <= 14 {
		return many
	}
	switch n % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return few
	default:
		return many
	}
}

func CountCodeLines(root string) int {
	var total int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		total += len(strings.Split(string(data), "\n"))
		return nil
	})

	if err != nil {
		total = 2000
		log.Print("services: failed to count code lines: %w", err)
	}

	return total
}
