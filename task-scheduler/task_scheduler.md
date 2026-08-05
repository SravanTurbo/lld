# Task Scheduler

## 1. Requirements

## Questions

- is scheduler responsible for task implementation - No

### Funcitonal

- store a task with future execution time.
- execute the task when time arrives.
- cancel a scheduled task before execution.

### Non-Functional

- in-memory
- thread-safe
- single process
- Derive the best achievable time complexity.

## 2. Algorithm

- Schedule(task, executeAt): register task
- Cancel(taskId): remove task
- Background Worker: pick the next executable task - execute task - remove task

## 3. State

### PerTask

- id
- command
- executeAt

### Scheduler

- scheduled tasks

## 4. Configuration

- questions:
- do we have a task capacity - No, whatever mem allows
- should same executeAt execute sequentially or in parallel: parallel
- should the execution order be guaranteed in the above case: No
- maximum retries: No retries

## 5. DS

### access patterns

- Schedule(task, executeAt):
  - insert
- Cancel(taskId):
  - lookup
  - remove
- PopNext():
  - find min timestamp item
  - remove

frequent: insert, findMin, remove
infrequent: lookup

### ds

- lookup can be made O(1) by Map
- remove can be made O(1) if we make it DLL
- findMin with min at head is O(1) - need to be sorted for every findMin to be O(1)
- insert is O(n) 

inserting in order is not important but min is important - min Heap solves this better
- lookup - O(1) with Map
- popMin - O(log n)
- insert - O(log n)
- peekMin - O(1)
- remove - O(log n)

### final
type task struct {
    id int
    command string
    executeAt time

    //Representation state
    heapIndex int
   }

- Scheduler
  - tasks: map[id]*task
  - minheap: []*task

## 6. Invariants

### domain

- A scheduled task exists until it is either execute or cancelled.
- A task is execute at or after scheduled time.
- A task is executed at most once.
- A cancelled task is never executed.

### representation

- count of items in map and heap should be same always
- heap property always holds (parent.executeAt <= children.executeAt>)
- task's heapIndex points to the tasts current position in the heap
- heap[task.heapIndex] = task


## 7. Concurrency

- one global lock on scheduler, to make sure tasks and heap are in sync
  - Scheduler
    - mu sync.lock
    - tasks: map[id]*task
    - minheap: []*task

## 8. Public API:

- NewScheduler()
- Schedule(task)
- Cacel(taskId)

## 9. Complexity
- Schedule - insert into map - O(1) + heap insert (sift up) - O(log n) 
- Cancel - lookup, delete from map - O(1) +  Heap remove (sift up/down) - O(log n)
- popNext - popmin (sift down) - O(log n) + remove from map - O(1)

overall O(log n) for all schedule, cancel, execute


to ask:
- should I ask all questions at the start? I dont think thats how software is built, we would discover new questions like at configuration