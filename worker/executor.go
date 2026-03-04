package worker

import (
	"dist-ocr/rpc"
	"fmt"
	"log"
	"time"
)

// Executor represents a Worker Node's processing engine.
// It holds the local TaskDeque and acts as the RPC receiver for the Master.
type Executor struct {
	ID    string
	Deque *TaskDeque
}

// NewExecutor creates a new worker engine.
func NewExecutor(id string) *Executor {
	return &Executor{
		ID:    id,
		Deque: NewTaskDeque(),
	}
}

// WorkerRPC is the wrapper struct used strictly for net/rpc registration.
// We expose this so that the RPC method signatures perfectly match what `net/rpc` expects:
// func (t *T) MethodName(argType T1, replyType *T2) error
type WorkerRPC struct {
	executor *Executor
}

// NewWorkerRPC creates the RPC-compatible wrapper.
func NewWorkerRPC(exec *Executor) *WorkerRPC {
	return &WorkerRPC{executor: exec}
}

// AssignTask is the RPC method called by the Master node to assign work.
func (w *WorkerRPC) AssignTask(req rpc.TaskRequest, resp *rpc.TaskResponse) error {
	log.Printf("[Worker %s] Received task %s from Job %s", w.executor.ID, req.TaskID, req.JobID)

	// 1. In a real system, we might push this to the Deque and return an ACK immediately,
	//    and let a background worker thread pop and process it asynchronously.
	// 2. However, for this MVP (and since we don't have async Wails callbacks for tasks yet),
	//    we will process it synchronously right here and return the result.
	
	// Push to deque (just to show we are using it during the task)
	w.executor.Deque.PushBottom(req)
	
	// Immediately pop it for processing
	poppedReq, ok := w.executor.Deque.PopBottom()
	if !ok {
		return fmt.Errorf("failed to retrieve task from local deque")
	}

	// ---------------------------------------------------------
	// MOCK TESSERACT OCR PROCESSING
	// ---------------------------------------------------------
	// Simulated processing delay
	time.Sleep(1 * time.Second)
	// Mock success Extraction
	mockText := fmt.Sprintf("MOCK_TEXT_FOR_PAGE_%d", poppedReq.PageNum)
	// ---------------------------------------------------------

	// Populate the response
	resp.TaskID = poppedReq.TaskID
	resp.WorkerID = w.executor.ID
	resp.ExtractedText = mockText
	
	log.Printf("[Worker %s] Completed task %s", w.executor.ID, req.TaskID)
	return nil
}
