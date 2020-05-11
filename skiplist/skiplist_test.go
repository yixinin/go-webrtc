package skiplist

import (
	"fmt"
	"testing"
)

func TestQuic(t *testing.T) {
	// quic.ListenAddr()
}

type IntValue int

func (s IntValue) Less(v Value) bool {
	return int(s) < int(v.Index())
}
func (s IntValue) Index() int {
	return int(s)
}

type UserValue struct {
	IntValue
	Name string
	Age  int
}

func TestSkipList(t *testing.T) {
	var list = NewSkipList()
	for i := 0; i < 20; i++ {
		var user = UserValue{
			Age:      i + 10,
			Name:     fmt.Sprintf("user%d", i),
			IntValue: IntValue(i),
		}

		list.Add(user)
		fmt.Println(list.List())
	}

}

func TestMountain(t *testing.T) {
	var arr = &MountainArray{
		arr: make([]int, 10),
	}
	for i := 0; i < 7; i++ {
		arr.arr[i] = i
	}
	for i := 9; i >= 7; i-- {
		arr.arr[i] = i - 5
	}
	findInMountainArray(10, arr)
	fmt.Println(arr.arr)
}

func TestEq(t *testing.T) {
	var n = equalSubstring("abcd", "bcde", 3)
	fmt.Println(n)
}
