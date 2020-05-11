package skiplist

import (
	"fmt"
	"math/rand"
	"time"
)

var R = rand.New(rand.NewSource(time.Now().UnixNano()))

type Value interface {
	Less(Value) bool
	Index() int
}

type NodeValue struct {
	Index int
	Data  interface{}
}

type Node struct {
	Value Value
	level int

	Left   *Node
	Right  *Node
	Top    *Node
	Bottom *Node
}

type Skiplist struct {
	Head       *Node
	totalLevel int
}

func NewSkipList() *Skiplist {
	return &Skiplist{
		totalLevel: 5,
	}
}

func (s *Skiplist) Search(target Value) bool {
	var now = s.Head
	if now == nil {
		return false
	}
	for now.Value.Less(target) {
		if now.Right != nil {
			now = now.Right
		} else {
			if now.level > 1 {
				now = now.Bottom
			}
		}
		if now.Value == target {
			return true
		}
	}
	return false
}

func (s *Skiplist) Add(num Value) {
	var now = s.Head

	//new head
	if s.Head == nil {
		s.Head = &Node{
			Value: num,
			level: s.totalLevel,
		}
		now = s.Head
		for i := s.totalLevel - 1; i > 0; i-- {

			node := &Node{
				Value: num,
				level: i,
			}
			node.Top = now
			now.Bottom = node

			now = node
		}
		return
	}

	if num.Less(s.Head.Value) {
		var oldHeadValue = s.Head.Value

		now.Value = num
		//将head.Value替换为新值
		for now.Bottom != nil {
			now = now.Bottom
			now.Value = num
		}

		var oldRight = now.Right

		//插入旧的value
		var node = &Node{
			Value: oldHeadValue,
			level: 1,
		}
		if oldRight != nil {
			node.Right = oldRight
			oldRight.Left = node
		}

		node.Left = now
		now.Right = node

		//上层
		for i := 2; i <= s.totalLevel; i++ {
			if !NeedNode() {
				break
			}
			var topNode = &Node{
				level: i,
				Value: oldHeadValue,
			}
			topNode.Bottom = node
			node.Top = topNode

			//回溯上一层
			if now.Top != nil {
				now = now.Top
				if now.Right != nil {
					var oldRight = now.Right
					topNode.Left = now
					topNode.Right = oldRight
					now.Right = topNode
				} else {
					topNode.Left = now
					now.Right = topNode
				}
			}

		}
		return
	}

	// find insert place
	for {
		if now.Right != nil {
			// now = now.Right
			if num.Less(now.Right.Value) {
				if now.Right.level == 1 {
					break
				} else {
					now = now.Bottom
					continue
				}

			}
			now = now.Right
		} else {
			if now.Bottom != nil {
				now = now.Bottom
			} else {
				break
			}
		}
	}

	//insert
	var oldRight = now.Right
	var node = &Node{
		level: 1,
		Value: num,
	}
	node.Left = now

	now.Right = node

	if oldRight != nil {
		node.Right = oldRight
		oldRight.Left = node
	}

	for i := 2; i <= s.totalLevel; i++ {
		if !NeedNode() {
			break
		}

		var topNode = &Node{
			level: i,
			Value: num,
		}
		topNode.Bottom = node
		node.Top = topNode

		//回溯上一层
		for {
			if now.level == s.totalLevel {
				break
			}
			if now.Top != nil {
				now = now.Top
				break
			} else {
				now = now.Left
			}
		}

		if now.Right != nil {
			var oldRight = now.Right
			topNode.Left = now
			topNode.Right = oldRight
			now.Right = topNode
		} else {
			topNode.Left = now
			now.Right = topNode
		}
		node = topNode
	}

}

func (s *Skiplist) Erase(num Value) bool {
	var now = s.Head
	if now == nil {
		return false
	}
	for {
		if now.Right == nil && now.Bottom == nil {
			return false
		}
		if now.Right != nil {
			now = now.Right
		} else {
			if now.Bottom != nil {
				now = now.Bottom
			}
		}

		if !now.Value.Less(num) && !num.Less(now.Value) {
			break
		}
	}

	//删除头
	if num == s.Head.Value {
		//将第二个元素升级为头
		for now.Bottom != nil {
			now = now.Bottom
		}
		//只有头 全部删除
		if now.Right == nil {
			s.Head = nil
			return true
		}
		var secondValue = now.Right.Value
		now.Value = secondValue
		var headNow = now
		for headNow.Top != nil {
			headNow = headNow.Top
			headNow.Value = secondValue
		}

		now = now.Right
	}

	for {
		if now == nil {
			return true
		}
		var left = now.Left
		var right = now.Right
		if right != nil {
			left.Right = right
			right.Left = left
		} else {
			left.Right = nil
		}

		now = now.Bottom
	}
}

func (s *Skiplist) List() []Value {
	var now = s.Head
	if now == nil {
		return nil
	}
	for now.Bottom != nil {
		now = now.Bottom
	}
	var arr = make([]Value, 0)
	for now != nil {
		arr = append(arr, now.Value)
		now = now.Right
	}
	return arr
}

func (s *Skiplist) Graph() {
	var head = s.Head
	if s.Head == nil {
		return
	}
	fmt.Println()
	for i := s.totalLevel; i > 0; i-- {
		var now = head
		var arr = make([]Value, 0)
		for now != nil {
			arr = append(arr, now.Value)
			now = now.Right
		}
		fmt.Println(arr)
		head = head.Bottom
	}
}

func NeedNode() bool {
	return R.Int()%2 == 0
}
