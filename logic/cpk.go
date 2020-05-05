package logic

import (
	"fmt"
	"math"
	"sort"
)

// RMSError 样本标准差计算公式
func RMSError(datas []float64) float64 {
	n := len(datas)
	if n == 0 {
		return 0
	}

	var sum = float64(0)
	for _, v := range datas {
		sum = sum + v
	}
	avg := sum / float64(n)

	var subPowSum float64
	for _, v := range datas {
		subPowSum = subPowSum + math.Pow(v-avg, 2)
	}

	return math.Sqrt(subPowSum / float64(n))
}

// Cp Cp计算公式
func Cp(tu, tl, s float64) float64 {
	t := tu - tl
	if s == 0 {
		return 0
	}
	r := t / (s * 6)
	return r
}

// Cpk 计算公式
func Cpk(tu, tl, u, s float64) float64 {
	if u == 0 || s == 0 {
		return 0
	}
	return math.Min(tu-u, u-tl) / (3 * s)
}

// Distribute 计算正太分布点
func Distribute(s, a float64, valueSet []float64) (min float64, max float64, values []float64, freqs []int, distribution []float64) {
	freqMap := make(map[float64]int)
	values = make([]float64, 0)
	freqs = make([]int, 0)
	distribution = make([]float64, 0)

	if len(valueSet) == 0 {
		return
	}

	max = valueSet[0]
	min = max

	for _, v := range valueSet {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}

		count := freqMap[v]
		count++
		freqMap[v] = count
	}

	sortedValues := make([]float64, 0)

	for k, _ := range freqMap {
		sortedValues = append(sortedValues, k)
	}

	sort.Float64s(sortedValues)

	for _, v := range sortedValues {
		values = append(values, v)
		freqs = append(freqs, freqMap[v])
		distribution = append(distribution, distributeFunc(s, a, v))
	}

	return
}

func distributeFunc(s, a, x float64) float64 {
	if s == 0 {
		return 0
	}
	part1 := 1 / (math.Sqrt(2*math.Pi) * s)
	part2 := math.Exp((-1 * (x - a) * (x - a)) / (2 * s * s))

	return math.Round(part1*part2*100) / 100
}
