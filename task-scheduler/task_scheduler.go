package taskscheduler

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

/*

1. FR:
- Schedule(task, executeAt)
- Cancel(taskId)
- Execute(task) @executeAt
NFR:
- in-memory
- thread safe
- single process
- best SLAs - O(log n)

2. Algo:
- schedule:
	- register task
- cancel:
	- lookup
	- Remove
- execute: background process that picks the task with nearest executeAt - what is the min executeAt?
	- task
	- Remove

3. State:
- task:
	- id
	- executeAt

- Scheduler:
	- min executeAt
	- tasks

4. Config:
- capacity?
- max retries?
- preferene for tasks at same executeAT
- parallel execution

5. DS:
- accesspatterns:
	- insert: O(log n)
	- lookup: items map - O(1)
	- Remove: O(log n)
	- next: O(log n)

- ds that can retain min best - minHeap, balanced BST

6. Concurrency: since state is mutated at once from itemsMap and minHeap, better to have a global lock on Scheduler

7.Invariants:
- task exists until it is executed or cancelled
- a task is executed at or after the schedule
- a task is executed at most once
- a cancelled task is never executed

Representation:
- Every task in minHeap exists in itemsMap.
- Every task in itemsMap exists exactly once in minHeap.
- task's heapIndex should point to the current index in minHeap

8. Public API:
- NewScheduler()
- Schedule(task)
- Cancel(taskId)

*/

type task struct {
	id int
	cmd string
	executeAt time.Time //UTC

	//representation state
	heapIndex int
}

type minHeap struct {
	storage []*task
}

func (mh *minHeap) Push(t *task) {
	t.heapIndex = len(mh.storage)

	mh.storage = append(mh.storage, t)
	mh.heapifyUp(t.heapIndex)
}

func (mh *minHeap) Peek() (*task, bool) {
	var t *task
	if len(mh.storage) == 0 {
		return t, false
	}

	t = mh.storage[0]
	return t, true
}

func (mh *minHeap) Pop() (*task, bool) {
	t, ok := mh.Peek()
	if !ok {
		return t, ok
	}

	mh.Remove(t)
	return t, true
}

func (mh *minHeap) Remove(t *task) {
	idx := t.heapIndex
	last := len(mh.storage)-1
	if idx == last {
		mh.storage = mh.storage[:last]
		t.heapIndex = -1
		return
	}

	mh.swap(idx, last)
	mh.storage = mh.storage[:last]
	t.heapIndex = -1

	if idx != 0 {
		parentIndex := (idx-1)/2
		curr := mh.storage[idx]
		parent := mh.storage[parentIndex]

		if curr.executeAt.Before(parent.executeAt){
			mh.heapifyUp(idx)
			return
		}
	}

	mh.heapifyDown(idx)
}

func (mh *minHeap) heapifyUp(idx int) {
	if idx == 0 {
		return
	}

	parentIndex := (idx - 1)/2

	curr := mh.storage[idx]
	parent := mh.storage[parentIndex]

	if curr.executeAt.After(parent.executeAt) || curr.executeAt.Equal(parent.executeAt) {
		return
	}

	mh.swap(idx, parentIndex)
	mh.heapifyUp(parentIndex)
}

func (mh *minHeap) swap(idx1, idx2 int) {
	temp := mh.storage[idx1]
	mh.storage[idx1] = mh.storage[idx2]
	mh.storage[idx2] = temp

	mh.storage[idx1].heapIndex = idx1
	mh.storage[idx2].heapIndex = idx2
}

func (mh *minHeap) heapifyDown(idx int) {
	childIdx, ok :=  mh.smallestChild(idx)
	if !ok {
		return
	}

	curr := mh.storage[idx]
	child := mh.storage[childIdx]

	if curr.executeAt.Before(child.executeAt) || curr.executeAt.Equal(child.executeAt) {
		return
	}

	mh.swap(idx, childIdx)
	mh.heapifyDown(childIdx)
}

func (mh *minHeap) smallestChild(idx int) (int, bool) {
	leftChildIdx := idx*2 + 1
	if leftChildIdx >= len(mh.storage) {
		return -1, false
	}

	rightChildIdx := idx*2 + 2
	if rightChildIdx >= len(mh.storage) {
		return leftChildIdx, true
	}

	if mh.storage[leftChildIdx].executeAt.Before(mh.storage[rightChildIdx].executeAt) {
		return leftChildIdx, true
	}

	return rightChildIdx, true
}

type Scheduler struct {
	mu sync.Mutex

	nextTaskID int
	items map[int]*task
	heap minHeap

	wakeUp chan struct{}
}

func NewScheduler() *Scheduler {
	s := Scheduler{
		items: make(map[int]*task),
		heap: minHeap{},
		wakeUp: make(chan struct{}, 1),
	}

	go s.start()
	return &s
}

func (s *Scheduler) start() {
	for {
		s.mu.Lock()
		t, ok := s.heap.Peek()
		s.mu.Unlock()

		if !ok {
			<-s.wakeUp
			continue
		}

		wait := time.Until(t.executeAt)
		if wait <= 0 {
			s.executeNext()
			continue
		}

		timer := time.NewTimer(wait)

		select {
		case <-s.wakeUp:
			timer.Stop() //resource cleanup
			s.executeNext()	
		case <-timer.C:
			s.executeNext() 
		} 
	}
}

func (s *Scheduler) run(t *task) {
	_, err := exec.Command(t.cmd).Output()
	if err != nil {
		fmt.Println("task failed: ", *t)		
	}
}

func (s *Scheduler) Schedule(t task) task {
	s.mu.Lock()

	next, ok := s.heap.Peek()

	s.nextTaskID++
	t.id = s.nextTaskID
	s.heap.Push(&t)
	s.items[t.id] = &t

	s.mu.Unlock()

	if !ok || next.executeAt.After(t.executeAt) {
		select {
			case s.wakeUp <- struct{}{}:
			default: //non-blocking if channel not ready
		}
	}

	return t
}

func (s *Scheduler) Cancel(taskId int) bool {
	s.mu.Lock()

	next, hasNext := s.heap.Peek()

	t, exists := s.items[taskId]
	if !exists {
		s.mu.Unlock()
		fmt.Println("no task to cancel")
		return false
	}

	delete(s.items, t.id)
	s.heap.Remove(t)

	s.mu.Unlock()

	if hasNext && t == next { 
		select {
		case s.wakeUp <- struct{}{}:
		default: //non-blocking if channel not ready
		}
	}	

	return true
}

func (s *Scheduler) executeNext() {
	s.mu.Lock()
	t, ok := s.heap.Peek()
	if !ok || time.Until(t.executeAt) > 0 {
		s.mu.Unlock()
		return
	}

	t, _ = s.heap.Pop()
	delete(s.items, t.id)

	s.mu.Unlock()

	go s.run(t)
}






