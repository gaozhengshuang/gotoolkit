/// @file rand.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-06-01

package util
import "math/rand"

// 百分比概率
func SelectPercent(per int32) bool {
	return SelectByOdds(per, 100)
}

// 千分比概率
func SelectPermillage(per int32) bool {
	return SelectByOdds(per, 1000)
}

// 万分比概率
func SelectTenThousand(per int32) bool {
	return SelectByOdds(per, 10000)
}

// 几分之几的概率
func SelectByOdds(per int32, max int32) bool {
	if per <= 0  { return false  }
	rd := rand.Int31n(max)	// a non-negative pseudo-random number in [0,max)
	if rd < per { return true   }
	return false
}

// 权重概率
func SelectByWeightSlice(vec []int32) int32 {
	totalweight := int32(0)
	for _, v := range vec { totalweight += v }

	count, weight := int32(0), RandBetween(1,totalweight);
	for k ,v := range vec {
		count += v
		if weight <= count { return int32(k) }
	}
	return -1
}

// 权重概率
type WeightOdds struct {
	Weight int32
	Uid	int64
	Num int64
}
func SelectByWeightOdds(vec []WeightOdds) int32 {
	totalweight := int32(0)
	for _, v := range vec { totalweight += v.Weight }

	count, weight := int32(0), RandBetween(1,totalweight);
	for k ,v := range vec {
		count += v.Weight
		if weight <= count { return int32(k) }
	}
	return -1
}



// 产生[min - max]内随机数，闭区间，例如 [-10 -- 20]
func RandBetween(min , max int32) int32 {
	if min == max { return min }
	if min > max  { panic("RandBetween min > max") }
	if min > max { max, min = min, max}
	diff := max - min + 1
	return rand.Int31n(diff) + min
}

func RandBetweenInt64(min , max int64) int64 {
	if min == max { return min }
	if min > max  { panic("RandBetween min > max") }
	diff := max - min + 1
	return rand.Int63n(diff) + min
}

// --------------------------------------------------------------------------
/// @brief 从指定范围total内，随机num个数据，不能重复
/// @param total 范围大小 [0 - total)
/// @param num 返回数据数量
///
/// @return 
// --------------------------------------------------------------------------
func SelectRandNumbers(total, num int32) []int32 {
	numbers := make([]int32, 0, total)
	for i := int32(0); i < total; i++ {
		numbers = append(numbers, i)
	}
	if num >= total {
		return numbers
	}

	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	return numbers[0:num]
}

