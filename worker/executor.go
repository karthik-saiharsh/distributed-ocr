package worker

import (
	"dist-ocr/rpc"
	"dist-ocr/swim"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Executor represents a Worker Node's processing engine.
// It holds the local TaskDeque and acts as the RPC receiver for the Master.
type Executor struct {
	ID    string
	Deque *TaskDeque

	swimSvc *swim.SWIMService
	rpcPort int

	mu      sync.Mutex
	pending map[string]chan rpc.TaskResponse

	ocrEngine *OCREngine
}

// NewExecutor creates a new worker engine.
func NewExecutor(id string, svc *swim.SWIMService, rpcPort int) *Executor {
	ocr, err := NewOCREngine()
	if err != nil {
		log.Printf("[Worker %s] Warning: OCR Engine failed to initialize: %v", id, err)
	}

	return &Executor{
		ID:        id,
		Deque:     NewTaskDeque(),
		swimSvc:   svc,
		rpcPort:   rpcPort,
		pending:   make(map[string]chan rpc.TaskResponse),
		ocrEngine: ocr,
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

	// Set VictimAddr to this node, so if it gets stolen, the thief knows where to return it.
	req.VictimAddr = fmt.Sprintf("%s:%d", w.executor.swimSvc.Self.IP, w.executor.rpcPort)

	// Create a channel to block on until the result is ready
	resCh := make(chan rpc.TaskResponse, 1)

	w.executor.mu.Lock()
	w.executor.pending[req.TaskID] = resCh
	w.executor.mu.Unlock()

	// Push to deque
	w.executor.Deque.PushBottom(req)

	// Block until someone (either local background worker or a thief) submits the result
	result := <-resCh

	w.executor.mu.Lock()
	delete(w.executor.pending, req.TaskID)
	w.executor.mu.Unlock()

	*resp = result
	return nil
}

// StealTask is called by another worker (thief) trying to steal tasks.
func (w *WorkerRPC) StealTask(req rpc.StealRequest, resp *rpc.StealResponse) error {
	stolen := w.executor.Deque.StealTop(req.Count)
	resp.Tasks = stolen
	if len(stolen) > 0 {
		log.Printf("[Worker %s] Stolen %d tasks by thief %s", w.executor.ID, len(stolen), req.ThiefID)
	}
	return nil
}

// SubmitStolenResult is called by a thief to return the processed result of a stolen task.
func (w *WorkerRPC) SubmitStolenResult(req rpc.TaskResponse, resp *bool) error {
	w.executor.mu.Lock()
	ch, ok := w.executor.pending[req.TaskID]
	w.executor.mu.Unlock()

	if ok {
		log.Printf("[Worker %s] Received result for stolen task %s from %s", w.executor.ID, req.TaskID, req.WorkerID)
		ch <- req
		*resp = true
	} else {
		log.Printf("[Worker %s] Received result for unknown/expired task %s", w.executor.ID, req.TaskID)
		*resp = false
	}
	return nil
}

// StartBackgroundWorker starts a loop that actively processes the local Deque
// or attempts to steal work if empty.
func (e *Executor) StartBackgroundWorker() {
	go func() {
		for {
			task, ok := e.Deque.PopBottom()

			if ok {
				// Process Task Locally
				e.processTask(task)
			} else {
				// Deque is empty, attempt to Steal Work!
				e.attemptSteal()
				time.Sleep(100 * time.Millisecond) // Don't thrash too hard if cluster is idle
			}
		}
	}()
}

func (e *Executor) processTask(task rpc.TaskRequest) {
	var extractedText string
	var execErr string

	if len(task.ImageData) > 0 {
		// 1. Save the incoming physical RPC byte slice to a unique temporary PNG file
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("worker_%s_task_%s.png", e.ID, task.TaskID))
		err := os.WriteFile(tempFile, task.ImageData, 0644)
		if err == nil {
			defer os.Remove(tempFile) // Ensure the temp image is cleaned up immediately after OCR

			// 2. Run real Tesseract CLI on the raw image
			if e.ocrEngine != nil {
				extractedText, err = e.ocrEngine.ExtractText(tempFile)
				if err != nil {
					execErr = err.Error()
				}
			} else {
				execErr = "OCR Engine not cleanly initialized on this worker (is Tesseract installed?)"
			}
		} else {
			execErr = fmt.Sprintf("Worker failed to write temp RPC image blob to disk: %v", err)
		}
	} else {
		// Graceful fallback for mock testing (or empty task payloads)
		time.Sleep(1 * time.Second)
		extractedText = fmt.Sprintf("MOCK_TEXT_FOR_PAGE_%d (Empty Payload)", task.PageNum)
	}

	resp := rpc.TaskResponse{
		TaskID:        task.TaskID,
		WorkerID:      e.ID,
		ExtractedText: extractedText,
		Error:         execErr,
	}

	// Check if this task belongs to us, or if we stole it
	selfAddr := fmt.Sprintf("%s:%d", e.swimSvc.Self.IP, e.rpcPort)
	if task.VictimAddr == selfAddr || task.VictimAddr == "" {
		// It's ours. Resolve it locally.
		e.mu.Lock()
		ch, ok := e.pending[task.TaskID]
		e.mu.Unlock()
		if ok {
			ch <- resp
		}
	} else {
		// We stole this! Send the result back to the victim.
		log.Printf("[Worker %s] Returning stolen result %s to victim %s", e.ID, task.TaskID, task.VictimAddr)
		client, err := rpc.NewClient(task.VictimAddr)
		if err != nil {
			log.Printf("[Worker %s] Failed to dial victim %s: %v", e.ID, task.VictimAddr, err)
			return
		}
		defer client.Close()

		_, err = client.SubmitStolenResult(resp)
		if err != nil {
			log.Printf("[Worker %s] Failed to submit stolen result to victim %s: %v", e.ID, task.VictimAddr, err)
		}
	}
}

func (e *Executor) attemptSteal() {
	nodes := e.swimSvc.GetClusterNodes()
	var alive []swim.Node
	for _, n := range nodes {
		if n.Status == swim.StatusAlive && n.ID != e.ID {
			alive = append(alive, n)
		}
	}

	if len(alive) == 0 {
		return // No one to steal from
	}

	// Pick a random victim
	victim := alive[rand.Intn(len(alive))]
	// Victim's RPC Port is their SWIM port + 1
	victimAddr := fmt.Sprintf("%s:%d", victim.IP, victim.Port+1)

	client, err := rpc.NewClient(victimAddr)
	if err != nil {
		return // Silent fail, the node might just be down.
	}
	defer client.Close()

	req := rpc.StealRequest{
		ThiefID: e.ID,
		Count:   1, // For MVP, steal 1 at a time
	}
	resp, err := client.StealTask(req)
	if err == nil && len(resp.Tasks) > 0 {
		log.Printf("[Worker %s] successfully stole %d tasks from %s", e.ID, len(resp.Tasks), victim.ID)
		for _, t := range resp.Tasks {
			e.Deque.PushBottom(t)
		}
	}
}
