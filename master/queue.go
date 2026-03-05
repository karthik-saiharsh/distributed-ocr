package master

import (
	"dist-ocr/rpc"
	"sync"
)

// Job represents an entire uploaded document.
type Job struct {
	ID         string
	TotalTasks int
	Tasks      []rpc.TaskRequest
}

// GlobalQueue holds all pending tasks across all jobs.
// It is thread-safe as both Wails frontend (adding jobs) and
// the dispatcher loop (removing tasks) will access it concurrently.
type GlobalQueue struct {
	mu          sync.Mutex
	pendingJobs map[string]*Job
	queue       []rpc.TaskRequest // FIFO queue of tasks awaiting assignment
}

func NewGlobalQueue() *GlobalQueue {
	return &GlobalQueue{
		pendingJobs: make(map[string]*Job),
		queue:       make([]rpc.TaskRequest, 0),
	}
}

// AddJob registers a new job and pushes all its fragmented tasks into the queue.
func (q *GlobalQueue) AddJob(job *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.pendingJobs[job.ID] = job
	// Push all tasks to the back of the queue
	q.queue = append(q.queue, job.Tasks...)
}

// PopTask retrieves the next task from the front of the queue.
// Returns false if the queue is empty.
func (q *GlobalQueue) PopTask() (rpc.TaskRequest, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return rpc.TaskRequest{}, false
	}

	task := q.queue[0]
	q.queue = q.queue[1:]
	return task, true
}

// RequeueTask puts a failed/mismatched task back at the front of the queue
// so it gets retried immediately.
func (q *GlobalQueue) RequeueTask(task rpc.TaskRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Prepend to the slice
	q.queue = append([]rpc.TaskRequest{task}, q.queue...)
}

// Len returns the number of pending tasks.
func (q *GlobalQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

// GetJob safely returns a Job by ID for metadata lookups (like total pages).
func (q *GlobalQueue) GetJob(jobID string) (*Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	job, exists := q.pendingJobs[jobID]
	return job, exists
}
