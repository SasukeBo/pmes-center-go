package logic

import (
	"fmt"
	"math"
)

// RMSError 样本标准差计算公式
func RMSError(datas []float64) float64 {
	n := len(datas)
	if n <= 1 {
		return 0
	}

	var sum = float64(0)
	for _, v := range datas {
		sum = sum + v
	}
	fmt.Println("总和为：", sum)
	fmt.Println("长度为：", n)
	avg := sum / float64(n)
	fmt.Println("平均值为：", avg)

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
