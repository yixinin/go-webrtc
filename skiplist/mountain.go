package skiplist

import "fmt"

type MountainArray struct {
	arr []int
}

func (a *MountainArray) get(i int) int {
	if i >= len(a.arr) {
		return -1
	}
	return a.arr[i]
}
func (a *MountainArray) length() int {
	return len(a.arr)
}

func findInMountainArray(target int, mountainArr *MountainArray) int {
	fmt.Println("peak:", findTop(mountainArr))
	return 0
}

func findTop(mountainArr *MountainArray) int {
	var left = 0
	var right = mountainArr.length() - 1
	var m, peak int
	for left < right {
		m = (left + right) / 2
		if mountainArr.get(m) < mountainArr.get(m+1) {
			left = m + 1
			peak = m + 1
		} else {
			right = m
		}
	}
	return peak
}
