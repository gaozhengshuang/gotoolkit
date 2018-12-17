/// @file container.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-04-01

package util
import (
	_"time"
	"math/rand"
)


// 打乱 slice numbers, 效率不如 rand.Shuffle
//func Shuffle(numbers []int32) {
//	for i := range numbers {
//		j := rand.Intn(i + 1)
//		numbers[i], numbers[j] = numbers[j], numbers[i]
//	}
//}

func Shuffle(numbers []int) {
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}

func ShuffleInt32(numbers []int32) {
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}

func ShuffleUint32(numbers []uint32) {
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}

func ShuffleInt64(numbers []int64) {
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}

func ShuffleUint64(numbers []uint64) {
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}

func ShuffleString(numbers []string) {
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
}



// --------------------------------------------------------------------------
/// @brief 从切片移除元素列表
/// @brief
///
/// @param []int32
/// @param []int32
///
/// @return 
// --------------------------------------------------------------------------
func RemoveSliceElemInt(values []int, discards []int) []int {
	rmIndexs := make([]int, 0)
	for i, v := range values {
		for _, d := range discards {
			if v != d { continue }
			rmIndexs = append(rmIndexs, i)
		}
	}

	for i, v := range rmIndexs {
		rmIndex := v - i
		values = append(values[0:rmIndex], values[rmIndex+1:]...)
	}

	return values
}

func RemoveSliceElemInt32(values []int32, discards []int32) []int32 {
	rmIndexs := make([]int, 0)
	for i, v := range values {
		for _, d := range discards {
			if v != d { continue }
			rmIndexs = append(rmIndexs, i)
		}
	}

	for i, v := range rmIndexs {
		rmIndex := v - i
		values = append(values[0:rmIndex], values[rmIndex+1:]...)
	}

	return values
}

func RemoveSliceElemUint32(values []uint32, discards []uint32) []uint32 {
	rmIndexs := make([]int, 0)
	for i, v := range values {
		for _, d := range discards {
			if v != d { continue }
			rmIndexs = append(rmIndexs, i)
		}
	}

	for i, v := range rmIndexs {
		rmIndex := v - i
		values = append(values[0:rmIndex], values[rmIndex+1:]...)
	}

	return values
}

func RemoveSliceElemInt64(values []int64, discards []int64) []int64 {
	rmIndexs := make([]int, 0)
	for i, v := range values {
		for _, d := range discards {
			if v != d { continue }
			rmIndexs = append(rmIndexs, i)
		}
	}

	for i, v := range rmIndexs {
		rmIndex := v - i
		values = append(values[0:rmIndex], values[rmIndex+1:]...)
	}

	return values
}

func RemoveSliceElemUint64(values []uint64, discards []uint64) []uint64 {
	rmIndexs := make([]int, 0)
	for i, v := range values {
		for _, d := range discards {
			if v != d { continue }
			rmIndexs = append(rmIndexs, i)
		}
	}

	for i, v := range rmIndexs {
		rmIndex := v - i
		values = append(values[0:rmIndex], values[rmIndex+1:]...)
	}

	return values
}

