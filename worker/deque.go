package worker

import (
	"dist-ocr/rpc"
	"sync"
)

// TaskDeque is a thread-safe double-ended queue for tasks.
// It acts as a Stack (LIFO) for the local worker to maximize cache locality,
// and as a Queue (FIFO) for work stealers to take the "coldest" tasks.
type TaskDeque struct {
	mu    sync.Mutex
	tasks []rpc.TaskRequest
}

// NewTaskDeque initializes an empty deque.
func NewTaskDeque() *TaskDeque {
	return &TaskDeque{
		tasks: make([]rpc.TaskRequest, 0),
	}
}

// PushBottom adds a new task to the local end (bottom) of the deque.
// Used when the Master assigns new work.
func (d *TaskDeque) PushBottom(task rpc.TaskRequest) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tasks = append(d.tasks, task)
}

// PopBottom removes and returns a task from the local end (bottom).
// Used by the local OCR worker thread to get its next piece of work.
// Returns false if the deque is empty.
func (d *TaskDeque) PopBottom() (rpc.TaskRequest, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	n := len(d.tasks)
	if n == 0 {
		return rpc.TaskRequest{}, false
	}

	task := d.tasks[n-1]
	d.tasks = d.tasks[:n-1]
	return task, true
}

// StealTop removes and returns up to 'count' tasks from the remote end (top).
// Used by other nodes performing Work Stealing.
// Returns the stolen tasks and the number actually stolen.
func (d *TaskDeque) StealTop(count int) []rpc.TaskRequest {
	d.mu.Lock()
	defer d.mu.Unlock()

	n := len(d.tasks)
	if n == 0 {
		return nil
	}

	if count > n {
		count = n
	}

	// Extract the tasks from the top (index 0)
	stolen := make([]rpc.TaskRequest, count)
	copy(stolen, d.tasks[:count])

	// Shift the remaining tasks down
	d.tasks = d.tasks[count:]

	return stolen
}

// Len returns the current number of tasks in the deque.
func (d *TaskDeque) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.tasks)
}
