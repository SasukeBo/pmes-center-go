package logic

func Normal(values []float64, freqs []int) float64 {
	total := 0
	for _, t := range freqs {
		total = total + t
	}
	if total == 0 {
		return 0
	}

	normal := float64(0)
	for i, v := range values {
		normal = normal + float64(freqs[i])/float64(total)*v
	}

	return normal
}
