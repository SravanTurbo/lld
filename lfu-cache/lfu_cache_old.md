# LFU

## 1. Problem Statement

Design a in-memory LFU cache

## 2.Requirements

### Questions/Assumptions

what changes the frequency?
assumption: only reads

### Functional Requirements

- Capacity based
- Put
- Get
- Put doesn't move
- Get moves away from LFU
- LFU eviction when full

### Non-Functional Requirements

- single process
- thread safe
- in-memory
- Time Complexity?

## 3. Algorithm

Existing Put: - Update

New Put: - Evict if full - Add item

Get: - LookUp - move away from LFU - Return value

## 4.State

Per Node: - key - value - frequency - recency

Cache: - capacity

## 5. Configuration

- capacity

## 6. Data Structure

### Access-Patterns

- Lookup By key: - Map at O(1)
- update value/freq - map with O(1)
- findPosition: - frequencyGroups with MRU at head - O(1)
- evict/insert - linkedlist makes it O(1), Map is O(1)

### ds

type Node struct{
    key string
    value string
    frequency int

    prev *Node
    next *Node
}

type frequencyGroup struct{
    head *Node
    tail *Node
}

type LFU struct{
    capacity int
    minFrequency int

    items Map[key]*Node
    groups Map[int]*frequencyGroup
}

## 7. Invariants

- minFrequency is the minimum frequency among all non-empty frequency groups
- item belongs to one and only frequency group
- Within a frequency group, items are ordered by recency
- count in items and groups should be same

## 8. Concurrency

- One global lock at LFU level - minFrequency, items and groups all are mutable in one Get
- sync.Mutex - both Get and Put change state

## 9. API

- Put(key, value)
- Get(key)

## 10.Implementation

- check lfu_cache.go
