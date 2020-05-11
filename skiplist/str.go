package skiplist

import "sort"

type IntSlice []int

func (a IntSlice) Len() int           { return len(a) }
func (a IntSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a IntSlice) Less(i, j int) bool { return a[i] < a[j] }

func equalSubstring(s string, t string, maxCost int) int {
	if len(s) != len(t) {
		return -1
	}
	if len(s) == 0 || len(t) == 0 {
		return -1
	}
	var arr = make(IntSlice, 0, len(s))

	for i := 0; i < len(s); i++ {
		var c = int(s[i]) - int(t[i])
		if c < 0 {
			c = -c
		}
		arr = append(arr, int(c))
	}

	sort.Sort(arr)
	var l = 0
	for i := 0; i < len(arr); i++ {
		if l+arr[i] > maxCost {
			return l
		}
		l += arr[i]
	}
	return 0
}

func maxScore(cardPoints []int, k int) int {
	var n = len(cardPoints)
	if k == 1 {
		if cardPoints[0] > cardPoints[n-1] {
			return cardPoints[0]
		}
		return cardPoints[n-1]
	}

	if k == n {
		var sum = 0
		for _, v := range cardPoints {
			sum += v
		}
		return sum
	}

	//重新组合
	var arr = make([]int, 0, k)
	for i := n - k; i < n; i++ {
		arr = append(arr, cardPoints[i])
	}

	for i := 0; i < k; i++ {
		arr = append(arr, cardPoints[i])
	}

	//取一个最大区间
	var maxSum = 0
	for i := 0; i < len(arr)-k; i++ {
		var sum = 0
		for j := 0; j < k; j++ {
			sum += arr[i+j]
		}
		if sum > maxSum {
			maxSum = sum
		}
	}
	return maxSum
}
