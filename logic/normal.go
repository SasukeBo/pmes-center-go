package logic

func solveAvg(values []float64) float64 {
	total := len(values)
	if total == 0 {
		return 0
	}

	sum := float64(0)
	for _, v := range values {
		sum = sum + v
	}

	avg := sum / float64(total)
	return avg
}
