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
	- remove
- execute: background process that picks the task with nearest executeAt - what is the min executeAt?
	- task
	- remove

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
	- remove: O(log n)
	- next: O(log n)

- ds that can retain min best - minHeap, balanced BST

6. Concurrency: since state is mutated at once from items and minHeap, better to have a global lock on Scheduler

7.Invariants:
- task exists until it is executed or cancelled
- a task is executed at or after the schedule
- a task is executed at most once
- a cancelled task is never executed

Representation:
- count in minHeap and items should always be equal
- an item in map should have an item in minHeap and vice-versa
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

func (mh *minHeap) add(t *task) {
	t.heapIndex = len(mh.storage)

	mh.storage = append(mh.storage, t)
	mh.siftUp(t.heapIndex)
}

func (mh *minHeap) siftUp(idx int) {
	parentIndex := (idx - 1)/2

	curr := mh.storage[idx]
	parent := mh.storage[parentIndex]

	if curr.executeAt.After(parent.executeAt) || curr.executeAt.Equal(parent.executeAt) {
		return
	}

	mh.swap(idx, parentIndex)
	mh.siftUp(parentIndex)
}

func (mh *minHeap) swap(idx1, idx2 int) {
	temp := mh.storage[idx1]
	mh.storage[idx1] = mh.storage[idx2]
	mh.storage[idx2] = temp

	mh.storage[idx1].heapIndex = idx2
	mh.storage[idx2].heapIndex = idx1
}

func (mh *minHeap) remove(t *task) {
	mh.swap(t.heapIndex, len(mh.storage)-1)
	mh.storage = mh.storage[:len(mh.storage)-1]
	
	parentIndex := (t.heapIndex-1)/2
	curr := mh.storage[t.heapIndex]
	parent := mh.storage[parentIndex]

	if curr.executeAt.Before(parent.executeAt){
		mh.siftUp(t.heapIndex)
	} else {
		mh.siftDown(t.heapIndex)
	} 
}

func (mh *minHeap) siftDown(idx int) {
	L := len(mh.storage)

	leftIdx := (2*idx) + 1
	rightIdx := (2*idx) + 2

	if leftIdx < L && mh.storage[leftIdx].executeAt.Before(mh.storage[idx].executeAt) {
		mh.swap(idx, leftIdx)
		mh.siftDown(leftIdx)
	} else if rightIdx < L && mh.storage[rightIdx].executeAt.Before(mh.storage[idx].executeAt) {
		mh.swap(idx, rightIdx)
		mh.siftDown(rightIdx)
	}

}

func (mh *minHeap) peek() (*task, bool) {
	var t *task
	if len(mh.storage) == 0 {
		return t, false
	}

	t = mh.storage[0]
	return t, true
}

type Scheduler struct {
	mu sync.Mutex

	count int
	items map[int]*task
	heap minHeap
}

func NewScheduler() *Scheduler {
	s := Scheduler{
		items: make(map[int]*task),
		heap: minHeap{},
	}

	s.start()
	return &s
}

func (s *Scheduler) start() {
	for true {
		t, ok := s.heap.peek()
		if !ok {
			time.Sleep(10000) //sleep 1m if no tasks -  this delays execution by 1m, we can reduce this delay using a channel and listening on new tasks
			continue
		}

		if time.Since(t.executeAt) >= 0 {
			s.execute(t)
		}else{
			time.Sleep(time.Until(t.executeAt))
		}
	}
}

func (s *Scheduler) run(t *task) {
	_, err := exec.Command(t.cmd).Output()
	if err != nil {
		fmt.Println("task failed: ", *t)		
	}
}

func (s *Scheduler) Schedule(cmd string, executeAt time.Time) task {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count++

	t := task{
		id: s.count,
		cmd: cmd,
		executeAt: executeAt,
	}

	s.heap.add(&t)
	s.items[t.id] = &t

	return t
}

func (s *Scheduler) Cancel(taskId int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.items[taskId]
	if !ok {
		fmt.Println("no task to cancel")
		return false
	}

	delete(s.items, t.id)
	s.heap.remove(t)

	return true
}

func (s *Scheduler) execute(t *task){
	s.mu.Lock()
	delete(s.items, t.id)
	s.heap.remove(t)
	s.mu.Unlock()

	s.run(t)
}





