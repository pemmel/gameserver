package game

import "sync"

func NewLList[T any]() *LlistHead[T] {
	return &LlistHead[T]{
		next: nil,
		tail: nil,
	}
}

type LlistHead[T any] struct {
	next  *LlistNode[T]
	tail  *LlistNode[T]
	mutex sync.Mutex // synchronization applies when chaining tail node
}

func (h *LlistHead[T]) Next() *LlistNode[T] {
	return h.next
}

func (h *LlistHead[T]) Borrow(n int, b *[]*LlistNode[T]) int {
	cnt := 0
	h.mutex.Lock()
	for i := h.next; i != nil && cnt < n; i, cnt = i.next, cnt+1 {
		*b = append(*b, i)
	}
	if cnt != 0 {
		tail := (*b)[len(*b)-1]
		if tail.next == nil {
			h.next = nil
			h.tail = nil
		} else {
			h.next = tail.next
			tail.next = nil
		}
	}
	h.mutex.Unlock()
	return cnt
}

func (h *LlistHead[T]) Return(head, tail *LlistNode[T]) {
	if head == nil || tail == nil {
		return
	}
	h.mutex.Lock()
	if h.tail == nil {
		h.next = head
		h.tail = tail
	} else {
		tail.next = h.next
		h.next = head
	}
	h.mutex.Unlock()
}

func (h *LlistHead[T]) Remove(pred func(*LlistNode[T]) bool) *LlistNode[T] {
	var p *LlistNode[T] = nil
	var i *LlistNode[T] = h.next
	for i != nil {
		if pred(i) {
			h.mutex.Lock()
			// there's scenario where i couldn't be removed:
			// 1. another thread just recently remove p (p.next is nil)
			// 2. another thread just recently remove i (p.next is i.next)
			if p.next != i {
				return nil
			} else if h.next == i {
				h.next = i.next
			} else {
				p.next = i.next
			}
			h.mutex.Unlock()
			i.next = nil
			return i
		}
		p = i
		i = i.next
	}
	return nil
}

func (h *LlistHead[T]) Insert(value T) {
	node := &LlistNode[T]{
		next:  nil,
		Value: value,
	}
	h.mutex.Lock()
	if h.tail == nil {
		h.next = node
		h.tail = node
	} else {
		h.tail.next = node
		h.tail = node
	}
	h.mutex.Unlock()
}

type LlistNode[T any] struct {
	next  *LlistNode[T]
	Value T
}

func (n *LlistNode[T]) Next() *LlistNode[T] {
	return n.next
}
