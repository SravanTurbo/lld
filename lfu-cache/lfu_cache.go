package lfucache

import "sync"

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
	minFrequency int

	items map[string]*Node
	groups map[int]*FrequencyGroup
}

func NewLFU(capacity int) *LFU{
	if capacity <= 0{
		panic("capacity must be >0")
	}

	return &LFU{
		capacity: capacity,
		minFrequency: 0,

		items: make(map[string]*Node),
		groups: make(map[int]*FrequencyGroup),
	}
}

func (cache *LFU) Get(key string) (string, bool){
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, ok := cache.items[key]
	if !ok{
		return "", false
	}

	//remove from current frequency group
	node.prev.next = node.next
	node.next.prev = node.prev

	node.frequency++

	cache.addToGroups(node)

	cache.updateMinFrequency(node.frequency-1)

	return node.value, true
}

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

	cache.add(node)

	cache.minFrequency = 0
}

func (cache *LFU) evict(){
	freqGroup, _ := cache.groups[cache.minFrequency]
	freqGroup.tail.prev.prev.next = freqGroup.tail.prev.next
	freqGroup.tail.prev = freqGroup.tail.prev.prev
}


func (cache *LFU) add(node *Node){
	cache.addToGroups(node)
	cache.addToItems(node)
}

func (cache *LFU)addToGroups(node *Node){
	freqGroup, ok := cache.groups[node.frequency]
	if !ok{
		freqGroup = cache.addGroup(node.frequency)
	}

	node.prev = freqGroup.head
	freqGroup.head.next.prev = node

	node.next = freqGroup.head.next
	freqGroup.head.next = node
}

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

func (cache *LFU)addToItems(node *Node){
	cache.items[node.key] = node
}

func (cache *LFU)updateMinFrequency(frequency int){
	if cache.minFrequency != frequency{
		return
	}

	minFreqGroup, _ := cache.groups[cache.minFrequency]
	if minFreqGroup.head.next != minFreqGroup.tail{
		return
	}

	cache.minFrequency++
}