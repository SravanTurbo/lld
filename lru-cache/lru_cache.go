// LRU Cache
// requirements
// algorithm
// state
// config: capacity
// datastructure: hashmap + DLL for O(1) Put and Get methods
// concurrency:sync.Mutex lock on LRUCache because Get is also a write
// API
// invariants:
//  - should not exceed capacity
//  - hashmap & DLL must always be in sync
//  - head and tail are always connected
//  - all nodes should have neighbours
//  - DLL order shoul represent recency
//  -

package lrucache

import "sync"

type Node struct{
	key string
	value string

	prev *Node
	next *Node
}

type LRUCache struct{
	mu sync.Mutex
	capacity int

	head *Node //MRU sentinel
	tail *Node //LRU sentinel

	cache map[string]*Node
}

func NewLRUCache(capacity int)*LRUCache{
	if capacity <= 0 {
		panic("capacity must be >0")
	}

	head := &Node{}
	tail := &Node{}

	tail.prev = head
	head.next = tail

	return &LRUCache{
		capacity: capacity,
		head: head,
		tail: tail,
		cache: make(map[string]*Node),
	}
}

func (lc *LRUCache)Put(key, value string){
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if node, ok := lc.cache[key]; ok {
		node.value = value
		lc.moveToFront(node)
	}else{
		if lc.capacity == len(lc.cache){
			nodeToEvict := lc.tail.prev
			delete(lc.cache, nodeToEvict.key)
			lc.remove(nodeToEvict)
		}

		node := &Node{key: key, value: value}
		lc.addToFront(node)

		lc.cache[key] = node
	}

}

func (lc *LRUCache)Get(key string) (string, bool){
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if node, ok := lc.cache[key]; ok{
		lc.moveToFront(node)
		return node.value, true
	}

	return "", false
}

func (lc *LRUCache) moveToFront(node *Node){
	lc.remove(node)
	lc.addToFront(node)
}

func (lc *LRUCache) addToFront(node *Node) {
	lc.head.next.prev = node
	node.prev = lc.head

	node.next = lc.head.next
	lc.head.next = node
}

func (lc *LRUCache) remove(node *Node) {
	if node == lc.head || node == lc.tail {
		return
	}

	node.prev.next = node.next
	node.next.prev = node.prev
}




