package merge

import linkedlist "github.com/phamnam2003/challenges/leetcode"

type ListNode linkedlist.LinkedList[int]

func MergeTwoLists(list1 *ListNode, list2 *ListNode) *ListNode {
	var head ListNode
	prev := &head

	for list1 != nil && list2 != nil {
		if list1.V < list2.V {
			prev.Next = (*linkedlist.LinkedList[int])(list1)
			list1 = (*ListNode)(list1.Next)
		} else {
			prev.Next = (*linkedlist.LinkedList[int])(list2)
			list2 = (*ListNode)(list2.Next)
		}

		prev = (*ListNode)(prev.Next)
	}

	if list1 != nil {
		prev.Next = (*linkedlist.LinkedList[int])(list1)
	} else {
		prev.Next = (*linkedlist.LinkedList[int])(list2)
	}

	return (*ListNode)(head.Next)
}
