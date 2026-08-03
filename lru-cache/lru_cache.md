# LRU Cache Design Journal

## 0. Design Journey

### Initial Thinking

Requirement: O(1) Get and Put with LRU eviction.

### Alternatives Explored

- Array → O(n) movement, rejected.
- HashMap only → no ordering, rejected.
- Linked List only → O(n) lookup, rejected.

Final choice: - HashMap + Doubly Linked List.

### State Evolution

Initially considered storing currentSize. Realised Go's len(map) is
O(1), so redundant state was removed.

### Concurrency Evolution

Initially considered RWMutex. Observation: Get() updates recency,
therefore it mutates shared state. Final choice: sync.Mutex.

### Invariants

Instead of reasoning about pointer assignments, we defined invariants
first: - Head/tail always connected. - DLL and map contain identical
nodes. - Head.next is MRU. - Tail.prev is LRU.

### Primitive Operations

Repeated pointer manipulation naturally became: - remove() -
addToFront() - moveToFront()

### Evolution Discussion

We discussed: - TTL - Expire-after-write vs expire-after-access -
Background cleanup - Why TTL requires a separate expiry index rather
than relying on LRU ordering.

------------------------------------------------------------------------

## 1. Problem Statement

Design an in-memory, thread-safe LRU cache supporting O(1) average Get
and Put.

## 2. Requirements

### Functional

- Fixed capacity
- Get
- Put
- Update moves to MRU
- Get moves to MRU
- Evict LRU when full

### Non-functional

- O(1) average operations
- Thread-safe
- Single process
- In-memory

## 3. Algorithm

Existing Put: - Update - Move to front

New Put: - Evict if full - Insert - Move to front

Get: - Lookup - Move to front - Return value

## 4. State

Per node: - key - value - recency

Cache: - capacity

## 5. Configuration

- capacity Future:
- TTL
- Cleanup interval
- Eviction callback

## 6. Data Structures

- map\[key\]\*Node
- Doubly Linked List

## 7. Concurrency

Shared mutable state: - map - DLL - head/tail

Lock: - sync.Mutex

## 8. Invariants

- Capacity never exceeded.
- Map and DLL contain same nodes.
- Head.next is MRU.
- Tail.prev is LRU.

## 9. Public API

- NewLRUCache()
- Get()
- Put()

## 10. Primitive Operations

- remove()
- addToFront()
- moveToFront()

## 11. Complexity

- Get O(1)
- Put O(1)
- Space O(capacity)

## 12. Evolution

- TTL
- Weighted capacity
- Metrics
- Sharding
- Distributed cache

## 13. Mistakes & Learnings

- Don't enforce capacity before checking existing key.
- Prefer invariants over pointer reasoning.
- Add new data structures for new access patterns.

## 14. Key Takeaways

Requirements → Operations → Data Structures → Concurrency → Invariants →
Implementation.
