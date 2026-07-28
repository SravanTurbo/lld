# Rate Limiter Design Journal (Fixed Window)

## 0. Design Journey

This section captures **how the design evolved**, not just the final
answer.

### Initial Thinking

-   Requirement: limit requests per user per time window.
-   Started with a simple map from `userID -> Counter`.

### First Concurrency Attempt

-   Considered protecting everything with a single mutex.
-   Problem: all users would contend on one lock.

**Learning:** Identify independent mutable state before choosing locks.

### Refinement

-   The users map is shared.
-   Each user's counter is independent.

Final locking strategy: - `sync.RWMutex` on the users map. -
`sync.Mutex` per `Counter`.

This allows requests for different users to proceed concurrently.

### Data Structure Evolution

Rejected: - Global counter (cannot isolate users) - Slice/Array (poor
lookup)

Chosen: - `map[userID]*Counter`

Reason: - O(1) lookup. - Store pointers because the next operation
mutates the counter.

### Evolution Discussion

We discussed how this design could evolve into: - Sliding Window - Token
Bucket - Leaky Bucket - Redis-backed distributed limiter - Cleanup of
inactive users

------------------------------------------------------------------------

## 1. Problem Statement

Design an in-memory, thread-safe Fixed Window Rate Limiter.

## 2. Requirements

### Functional

-   Limit requests per user.
-   Configurable window.
-   Reject requests beyond limit.

### Non-functional

-   O(1) average request handling
-   Thread-safe
-   Single process
-   In-memory

## 3. Algorithm

1.  Lookup user.
2.  Create counter if absent.
3.  Reset window if expired.
4.  Check limit.
5.  Increment and allow or reject.

## 4. State

### Per User

-   count
-   windowStart

### Global

-   limit
-   windowSize

## 5. Configuration

-   limit
-   windowSize

Future: - cleanup interval - idle TTL

## 6. Data Structures

-   `map[userID]*Counter`

## 7. Concurrency

Shared mutable state: - users map - counters

Strategy: - RWMutex for map - Mutex per counter

## 8. Invariants

-   One counter per user.
-   Counter belongs to one active window.
-   Allowed request never exceeds limit.
-   Map references valid counters.

## 9. Public API

-   NewRateLimiter()
-   Allow()

## 10. Primitive Operations

-   getOrCreateCounter()
-   resetWindowIfExpired()
-   incrementIfAllowed()

## 11. Implementation Summary

Lookup → Create if needed → Reset window → Check limit →
Increment/Reject

## 12. Complexity

-   Allow(): O(1) average
-   Space: O(active users)

## 13. Evolution

-   Sliding Window
-   Token Bucket
-   Leaky Bucket
-   Distributed limiter
-   Cleanup strategies

## 14. Mistakes & Learnings

-   Started with coarse locking.
-   Learnt to derive locking from shared mutable state.
-   Store pointers when subsequent operations mutate state.

## 15. Key Takeaways

1.  Requirements drive algorithm.
2.  Shared mutable state drives concurrency.
3.  Derive data structures from operations.
4.  Keep APIs small with primitive operations.
5.  Design for evolution without over-engineering.
