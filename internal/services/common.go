package services

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
