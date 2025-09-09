package reverse

import linkedlist "github.com/phamnam2003/challenges/leetcode"

type ListNode linkedlist.LinkedList[int]

func ReverseList(head *ListNode) *ListNode {
	var prev *ListNode
	for head != nil {
		next := head.Next
		head.Next = (*linkedlist.LinkedList[int])(prev)
		prev = head
		head = (*ListNode)(next)
	}

	return prev
}
