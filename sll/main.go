package sll

import (
	"fmt"
	"sort"
	"sync"
)

// Sll is a scored linked list.
type Sll struct {
	sync.RWMutex
	head   *Node
	tail   *Node
	scores nodeScoreList
}

// Node is a scored linked list node.
type Node struct {
	sync.Mutex
	Value interface{}
	Score uint64
	Next  *Node
	Prev  *Node
	// Might add a create field for TTL.
}

// New creates a new *Sll
func New() *Sll {
	return &Sll{
		scores: nodeScoreList{},
	}
}

// nodeScoreList holds a slice of *Node
// for sorting by score.
type nodeScoreList []*Node

// Read returns a *Node Value and increments the score.
func (n *Node) Read() interface{} {
	n.Lock()
	n.Score++
	n.Unlock()
	return n.Value
}

// nodeScoreList methods to satisfy the sort interface.

func (nsl nodeScoreList) Len() int {
	return len(nsl)
}

func (nsl nodeScoreList) Less(i, j int) bool {
	return nsl[i].Score < nsl[j].Score
}

func (nsl nodeScoreList) Swap(i, j int) {
	nsl[i], nsl[j] = nsl[j], nsl[i]
}

// Len returns the count of nodes in the *Sll.
func (ll *Sll) Len() uint {
	ll.RLock()
	defer ll.RUnlock()
	return uint(len(ll.scores))
}

// Head returns the head *Node.
func (ll *Sll) Head() *Node {
	ll.RLock()
	defer ll.RUnlock()
	return ll.head
}

// Tail returns the head *Node.
func (ll *Sll) Tail() *Node {
	ll.RLock()
	defer ll.RUnlock()
	return ll.tail
}

// HighScores takes an integer and returns the
// respective number of *Nodes with the higest scores
// sorted in ascending order. The last element will
// be the node with the highest score. Calling HighScores
// locks the *Sll for the duration of a binary sort
// of roughly O(log ll.Len()) time.
func (ll *Sll) HighScores(r int) nodeScoreList {
	ll.Lock()
	defer ll.Unlock()

	sort.Sort(ll.scores)
	// Return what's available
	// if more is being requested
	// than exists.
	if r > len(ll.scores) {
		scores := make(nodeScoreList, len(ll.scores))
		copy(scores, ll.scores)
		return scores
	}

	scores := make(nodeScoreList, r)
	copy(scores, ll.scores[len(ll.scores)-r:])

	return scores
}

// LowScores takes an integer and returns the
// respective number of *Nodes with the lowest scores
// sorted in ascending order. The first element will
// be the node with the lowest score. Calling LowScores
// locks the *Sll for the duration of a binary sort
// of roughly O(log ll.Len()) time.
func (ll *Sll) LowScores(r int) nodeScoreList {
	ll.Lock()
	defer ll.Unlock()

	sort.Sort(ll.scores)
	// Return what's available
	// if more is being requested
	// than exists.
	if r > len(ll.scores) {
		scores := make(nodeScoreList, len(ll.scores))
		copy(scores, ll.scores)
		return scores
	}
	scores := make(nodeScoreList, r)
	copy(scores, ll.scores[:r])
	return scores
}

// MoveToHead takes a *Node and moves it
// to the front of the *Sll.
func (ll *Sll) MoveToHead(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	// Short-circuit.
	if ll.head == n {
		return
	}

	// If this is the tail,
	// assign a new tail.
	if ll.tail == n {
		ll.tail = n.Next
	}

	// Set current head Next to n.
	ll.head.Next = n
	// Set n prev to current head.
	n.Prev = ll.head

	// Ensure n Next is nil.
	n.Next = nil
	// Set n to head.
	ll.head = n
}

// MoveToTail takes a *Node and moves it
// to the back of the *Sll.
func (ll *Sll) MoveToTail(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	// Short-circuit.
	if ll.tail == n {
		return
	}

	// If this is the head,
	// assign a new head.
	if ll.head == n {
		ll.head = n.Prev
	}

	// Set current tail Prev to n.
	ll.tail.Prev = n
	// Set n next to current tail.
	n.Next = ll.tail

	// Ensure n Prev is nil.
	n.Prev = nil
	// Set n to tail.
	ll.tail = n
}

// PushHead creates a *Node with value v
// at the head of the *Sll and returns a *Node.
func (ll *Sll) PushHead(v interface{}) *Node {
	ll.Lock()
	defer ll.Unlock()

	n := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, n)

	// Is this a new ll?
	if ll.head == nil {
		ll.head = n
		ll.tail = n
		return n
	}

	// Set current head next to n.
	ll.head.Next = n
	// Set n prev to current head.
	n.Prev = ll.head
	// Swap n to head.
	ll.head = n

	return n
}

// PushTail creates a *Node with value v
// at the tail of the *Sll and returns a *Node.
func (ll *Sll) PushTail(v interface{}) *Node {
	ll.Lock()
	defer ll.Unlock()

	n := &Node{
		Value: v,
		Score: 0,
	}

	ll.scores = append(ll.scores, n)

	// Is this a new ll?
	if ll.tail == nil {
		ll.head = n
		ll.tail = n
		return n
	}

	// Set current tail prev to n.
	ll.tail.Prev = n
	// Set n next to current tail.
	n.Next = ll.tail
	// Swap n to tail.
	ll.tail = n

	return n
}

// Remove removes a *Node from the *Sll.
func (ll *Sll) Remove(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	// If this is a single element list.
	if ll.head == n && ll.tail == n {
		ll.head, ll.tail = nil, nil
		// These should already be nil:
		n.Next, n.Prev = nil, nil
		goto updatescores
	}

	// Check if this node is the head/tail.
	// If head, promote prev to head.
	if ll.head == n {
		ll.head = n.Prev
		ll.head.Next = nil
		goto updatescores
	}
	// If tail, promote next to tail.
	if ll.tail == n {
		ll.tail = n.Next
		ll.tail.Prev = nil
		goto updatescores
	}

	// This node is otherwise at a non-end.
	// Link the next node and the prev.
	// TODO these used to have
	// if !nil checks; making it here
	// with nil should be considered a bug.
	n.Next.Prev = n.Prev
	n.Prev.Next = n.Next

updatescores:
	// Remove references.
	n.Next, n.Prev = nil, nil

	// Remove from score list.
	// Get index, remove element.
	//i := sort.Search(len(ll.scores), func(i int) bool { return ll.scores[i] == n })
	//fmt.Println(i)
	var i int
	for pos, node := range ll.scores {
		if node == n {
			i = pos
			break
		}
	}
	fmt.Printf("removing %v, found %v\n", n.Value, ll.scores[i].Value)
	if i == len(ll.scores) {
		// If the index is at the tail,
		// we just exclude the last element.
		ll.scores = ll.scores[:len(ll.scores)]
		return
	}

	ll.scores = append(ll.scores[:i], ll.scores[i+1:]...)

}

// RemoveHead removes the current *Sll.head.
func (ll *Sll) RemoveHead() {
	ll.Lock()
	defer ll.Unlock()

	if ll.head == nil {
		return
	}

	oldHead := ll.head

	// Set head to current head.Prev.
	ll.head = oldHead.Prev
	// Set new head Next to nil.
	ll.head.Next = nil
	// Remove old head references.
	oldHead.Next, oldHead.Prev = nil, nil

	var i int
	for pos, node := range ll.scores {
		if node == oldHead {
			i = pos
			break
		}
	}
	if i == len(ll.scores) {
		// If the index is at the tail,
		// we just exclude the last element.
		ll.scores = ll.scores[:len(ll.scores)]
		return
	}

	ll.scores = append(ll.scores[:i], ll.scores[i+1:]...)
}

// RemoveTail removes the current *Sll.tail.s
func (ll *Sll) RemoveTail() {
	ll.Lock()
	defer ll.Unlock()

	if ll.tail == nil {
		return
	}

	oldTail := ll.tail

	// Set tail to current tail.Next.
	ll.tail = oldTail.Next
	// Set new tail Prev to nil.
	ll.tail.Prev = nil
	// Remove old head references.
	oldTail.Next, oldTail.Prev = nil, nil

	var i int
	for pos, node := range ll.scores {
		if node == oldTail {
			i = pos
			break
		}
	}
	if i == len(ll.scores) {
		// If the index is at the tail,
		// we just exclude the last element.
		ll.scores = ll.scores[:len(ll.scores)]
		return
	}

	ll.scores = append(ll.scores[:i], ll.scores[i+1:]...)
}

// Special methods.

// PushHeadNode takes an existing *Node and
// sets it as the head of the *Sll. The *Node
// is also added to the score list.
func (ll *Sll) PushHeadNode(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	ll.scores = append(ll.scores, n)

	// Is this a new ll?
	if ll.head == nil {
		ll.head = n
		ll.tail = n
		n.Next, n.Prev = nil, nil
		return
	}

	// Set current head next to n.
	ll.head.Next = n
	// Set n prev to current head.
	n.Prev = ll.head
	// Swap n to head.
	ll.head = n
	// Set n Next to nil.
	n.Next = nil
}

// PushTailNode takes an existing *Node and
// sets it as the tail of the *Sll. The *Node
// is also added to the score list.
func (ll *Sll) PushTailNode(n *Node) {
	ll.Lock()
	defer ll.Unlock()

	ll.scores = append(ll.scores, n)

	// Is this a new ll?
	if ll.tail == nil {
		ll.head = n
		ll.tail = n
		n.Next, n.Prev = nil, nil
		return
	}

	// Set current tail prev to n.
	ll.tail.Prev = n
	// Set n next to current tail.
	n.Next = ll.tail
	// Swap n to tail.
	ll.tail = n
	// Set n Prev to nil.
	n.Prev = nil
}
