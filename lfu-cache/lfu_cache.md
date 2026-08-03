# LFU (Least Frequently Used Cache)

## 1. Problem Statement

Design a thread-safe, in-memory LFU (Least Frequently Used) cache.

When the cache reaches its capacity, evict the least frequently used
item. If multiple items have the same frequency, evict the least
recently used item among them.

## 2. Requirements

### Questions / Assumptions

  Question                                    Answer
  ------------------------------------------- --------------
  What increases frequency?                   Only `Get()`
  Does existing `Put()` increase frequency?   No
  Persistence required?                       No
  Single process?                             Yes

### Functional Requirements

-   Fixed capacity
-   Put(key, value)
-   Get(key)
-   Existing Put updates value only
-   New Put inserts item
-   LFU eviction
-   LRU tie-breaker

### Non-functional Requirements

-   In-memory
-   Thread-safe
-   Single process
-   O(1) Get
-   O(1) Put

## 3. Algorithm

### Put(existing)

-   Lookup
-   Update value

### Put(new)

-   Evict if full
-   Create node
-   Insert into frequency group 0
-   Register item
-   Update minFrequency

### Get

-   Lookup
-   Remove from current group
-   Increase frequency
-   Insert into next group
-   Update minFrequency if needed
-   Return value

## 4. State

### Per Node

-   key
-   value
-   frequency

### Cache

-   capacity
-   minFrequency

## 5. Configuration

-   capacity
-   score

## 6. Data Structure

-   items: map\[key\]\*Node
-   groups: map\[frequency\]\*FrequencyGroup
-   FrequencyGroup: Doubly linked list ordered MRU → LRU

## 7. Invariants

### Domain

-   Frequency never decreases.
-   Get increases frequency.
-   Existing Put does not.
-   Capacity never exceeded.
-   LFU eviction.
-   LRU tie-break.

### Representation

-   minFrequency is the minimum non-empty frequency.
-   Every node belongs to exactly one group.
-   Node.frequency equals its group's frequency.
-   Every node in items exists in exactly one group.
-   Empty groups are removed.

## 8. Concurrency

-   Single sync.Mutex.
-   Get mutates shared state.

## 9. Public API

-   Put(key, value)
-   Get(key)

## 10. Primitive Operations

-   addToFront
-   remove
-   popLRU
-   isEmpty
-   insertToGroup
-   registerItem

## 11. Complexity

  Operation        Time   Aux Space
  ---------------- ------ -----------
  Get              O(1)   O(1)
  Put(existing)    O(1)   O(1)
  Put(new)         O(1)   O(1)
  Put(new+evict)   O(1)   O(1)
  Evict            O(1)   O(1)

Overall Space: O(n)

## 12. Evolution

### Q1. How do we prevent old popular items from dominating forever?

**Answer**

Introduce frequency aging (decay or time-weighted frequency). TTL
removes stale entries but does not reduce historical popularity.

### Q2. How would you improve concurrency?

**Answer**

Current design uses a single mutex because Get mutates state. Higher
throughput requires a different representation or partitioned locking.

### Q3. What if there are millions of keys and frequency groups?

**Answer**

Memory, allocator and GC overhead increase. Consider aging, frequency
compression or capacity tuning.

### Q4. What if the process crashes?

**Answer**

Current implementation is volatile. Persist state only if recovery is a
requirement.

### Q5. How would you distribute the cache?

**Answer**

Local state becomes inconsistent. Options include sticky routing,
centralized/sharded metadata or approximate LFU.

### Q6. How would the API evolve?

**Answer**

Possible APIs: Delete, Peek, Clear.

Delete reveals that updating minFrequency in O(1) is no longer
guaranteed with the current representation.

## Key Learnings

-   Derive data structures from access patterns.
-   Separate domain invariants from representation invariants.
-   Encapsulate pointer manipulation behind primitive operations.
-   Evolution exposes hidden assumptions.
