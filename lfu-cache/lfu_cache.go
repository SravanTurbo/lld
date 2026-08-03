package lfucache

import (
	"sync"
)

type Node struct {
	key string
	value string
	frequency int

	prev *Node
	next *Node
}

type FrequencyGroup struct{
	head *Node //MFU sentinel
	tail *Node //LFU sentinel
}

type LFU struct {
	mu sync.Mutex
	capacity int
	score int
	minFrequency int

	items map[string]*Node
	groups map[int]*FrequencyGroup
}

func NewLFUCache(capacity int) *LFU{
	if capacity <= 0{
		panic("capacity must be >0")
	}

	return &LFU{
		capacity: capacity,
		score: 1, 
		minFrequency: 0,

		items: make(map[string]*Node),
		groups: make(map[int]*FrequencyGroup),
	}
}

//Time-Complexity: O(1)
func (cache *LFU) Put(key, value string){
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if node, ok := cache.items[key]; ok{
		node.value = value
		return
	}

	if cache.capacity == len(cache.items){
		cache.evict()
	}

	node := &Node{
		key: key,
		value: value,
		frequency: 0,
	}

	cache.insertToGroup(node)
	cache.items[node.key] = node
	cache.minFrequency = node.frequency
}

//Time-Complexity: O(1)
func (cache *LFU) evict(){
	freqGroup, ok := cache.groups[cache.minFrequency]
	if !ok{
		panic("cannot evict non-existing list")
	}

	lruItem := freqGroup.popLRU()

	if freqGroup.isEmpty(){
		delete(cache.groups, cache.minFrequency)
	}
	
	delete(cache.items, lruItem.key)
}

//Time-Complexity: O(1)
func (cache *LFU) insertToGroup(node *Node){
	freqGroup, ok := cache.groups[node.frequency]
	if !ok{
		freqGroup = cache.addGroup(node.frequency)
	}

	freqGroup.addToFront(node)
}

//Time-Complexity: O(1)
func (cache *LFU)addGroup(frequency int) *FrequencyGroup{
	head := &Node{}
	tail := &Node{}

	head.next = tail
	tail.prev = head

	cache.groups[frequency] = &FrequencyGroup{
		head: head,
		tail: tail,
	}

	return cache.groups[frequency]
}

//time-Complexity: O(1); space-complexity: O()
func (cache *LFU) Get(key string) (string, bool){
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, ok := cache.items[key]
	if !ok{
		return "", false
	}

	cache.promote(node)

	return node.value, true
}

//Time-Complexity: O(1)
func (cache *LFU) promote(node *Node){
	freqGroup, ok := cache.groups[node.frequency]
	if !ok{
		panic("frequency group not found, items and groups diverged")
	}

	freqGroup.remove(node)

	if freqGroup.isEmpty(){
		delete(cache.groups, node.frequency)
		cache.minFrequency = node.frequency + cache.score
	}

	node.frequency += cache.score

	cache.insertToGroup(node)
}

//Time-Complexity: O(1)
func (fg *FrequencyGroup) popLRU() *Node {
	if fg.isEmpty(){
		panic("cannot pop from empty frequency group")
	}

	lruItem := fg.tail.prev
	fg.remove(lruItem)
	return lruItem
}

//Time-Complexity: O(1)
func (fg *FrequencyGroup) isEmpty() bool {
	return fg.head.next == fg.tail
}

//Time-Complexity: O(1)
func (fg *FrequencyGroup) remove(node *Node){
	node.prev.next = node.next
	node.next.prev = node.prev

	node.next = nil
	node.prev = nil
}

//Time-Complexity: O(1)
func (fg *FrequencyGroup)addToFront(node *Node){
	node.prev = fg.head
	fg.head.next.prev = node

	node.next = fg.head.next
	fg.head.next = node
}

