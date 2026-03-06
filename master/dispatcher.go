package master

import (
	"context"
	"dist-ocr/rpc"
	"dist-ocr/swim"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Dispatcher constantly monitors the GlobalQueue and pushes tasks to workers via RPC.
type Dispatcher struct {
	Queue     *GlobalQueue
	Consensus *ConsensusEngine
	cluster   *swim.SWIMService
	stopCh    chan struct{}

	// Added to allow emitting the final result back to the frontend
	ctx context.Context

	// Track emitted jobs to guarantee single-fire payload
	emittedJobs map[string]bool
	emitMu      sync.Mutex

	// RPC Port for connecting to workers
	rpcPort int

	// Application-layer heartbeat tracking
	hbMu             sync.RWMutex
	workerHeartbeats map[string]time.Time
}

func NewDispatcher(cluster *swim.SWIMService, rpcPort int) *Dispatcher {
	return &Dispatcher{
		Queue:            NewGlobalQueue(),
		Consensus:        NewConsensusEngine(),
		cluster:          cluster,
		stopCh:           make(chan struct{}),
		rpcPort:          rpcPort,
		emittedJobs:      make(map[string]bool),
		workerHeartbeats: make(map[string]time.Time),
	}
}

// MasterRPC is the RPC wrapper for the Master node
type MasterRPC struct {
	dispatcher *Dispatcher
}

// NewMasterRPC creates the RPC-compatible wrapper.
func NewMasterRPC(disp *Dispatcher) *MasterRPC {
	return &MasterRPC{dispatcher: disp}
}

// Heartbeat records a keep-alive ping from a worker
func (m *MasterRPC) Heartbeat(req rpc.HeartbeatRequest, resp *bool) error {
	m.dispatcher.hbMu.Lock()
	m.dispatcher.workerHeartbeats[req.WorkerID] = time.Now()
	m.dispatcher.hbMu.Unlock()
	*resp = true
	return nil
}

// Start begins the background dispatch loop.
func (d *Dispatcher) Start(ctx context.Context) {
	d.ctx = ctx
	go d.dispatchLoop()
	log.Println("[Master] Dispatcher started.")
}

// Stop halts the dispatch loop.
func (d *Dispatcher) Stop() {
	close(d.stopCh)
}

func (d *Dispatcher) dispatchLoop() {
	ticker := time.NewTicker(200 * time.Millisecond) // Poll queue every 200ms
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			// If queue is empty, do nothing
			if d.Queue.Len() == 0 {
				continue
			}

			// Get active (Alive) nodes from SWIM
			allNodes := d.cluster.GetClusterNodes()
			var aliveNodes []swim.Node
			d.hbMu.RLock()
			now := time.Now()
			for _, n := range allNodes {
				eligible := false
				reason := ""

				if n.Status == swim.StatusAlive {
					// The local node (Master) doesn't send heartbeats to itself, so we must explicitly include it.
					if n.ID == d.cluster.Self.ID {
						aliveNodes = append(aliveNodes, n)
						eligible = true
						reason = "Local Master node implicitly eligible"
					} else {
						// Check application-layer heartbeat
						lastHb, ok := d.workerHeartbeats[n.ID]
						if ok && now.Sub(lastHb) <= 1500*time.Millisecond {
							aliveNodes = append(aliveNodes, n)
							eligible = true
							reason = "Recent app-layer heartbeat received"
						} else if ok {
							reason = "Missed app-layer heartbeats (stale)"
						} else {
							reason = "No app-layer heartbeat received yet"
						}
					}
				} else {
					reason = string(n.Status)
				}

				addr := fmt.Sprintf("%s:%d", n.IP, n.Port)
				log.Printf("[Master Debug] NodeID: %s, Address: %s, State: %s -> Eligible: %v (%s)", n.ID, addr, n.Status, eligible, reason)
			}
			d.hbMu.RUnlock()

			// DEBUG: Print state when queue has items
			log.Printf("[Master] Dispatch loop running. Queue Size: %d, Alive Nodes: %d", d.Queue.Len(), len(aliveNodes))

			// We only require at least 1 available worker.
			if len(aliveNodes) < 1 {
				log.Printf("[Master] Dropping task dispatch. Required: 1 nodes. Found: %d", len(aliveNodes))
				continue
			}

			// Grab a task
			task, ok := d.Queue.PopTask()
			if !ok {
				continue
			}

			var w1, w2 swim.Node
			if len(aliveNodes) >= 2 {
				// Pick two distinct random workers to assign this same task to
				w1, w2 = pickTwoDistinct(aliveNodes)
			} else {
				// Fallback to sending redundant tasks to the same available worker
				w1, w2 = aliveNodes[0], aliveNodes[0]
			}

			// Dispatch asynchronously
			go d.assignAndVerify(task, w1)
			go d.assignAndVerify(task, w2)
		}
	}
}

// assignAndVerify connects to a worker via RPC, sends the task, waits for response, and submits to Consensus.
func (d *Dispatcher) assignAndVerify(task rpc.TaskRequest, worker swim.Node) {
	// The worker's RPC port is always their SWIM port + 1 based on app.go assignments.
	workerRPCPort := worker.Port + 1
	addr := fmt.Sprintf("%s:%d", worker.IP, workerRPCPort) // Connect to the Worker's RPC port

	client, err := rpc.NewClient(addr)
	if err != nil {
		log.Printf("[Master] Failed to dial worker %s via %s: %v. Requeueing task %s", worker.ID, addr, err, task.TaskID)
		d.Queue.RequeueTask(task)
		return
	}
	defer client.Close()

	log.Printf("[Master] Assigning Task %s -> Worker %s", task.TaskID, worker.ID)

	// Make the blocking RPC Call
	resp, err := client.AssignTask(task)
	if err != nil {
		log.Printf("[Master] Task RPC failed on worker %s: %v. Requeueing %s", worker.ID, err, task.TaskID)
		d.Queue.RequeueTask(task)
		return
	}

	// Submit success to consensus engine
	job, exists := d.Queue.GetJob(task.JobID)
	if !exists {
		return // Job was cancelled or totally missing
	}

	ok, err := d.Consensus.SubmitResult(*resp, task.PageNum, task.JobID, job.TotalTasks)
	if err != nil {
		// Mismatch! The consensus engine tells us to requeue
		log.Printf("[Master] Consensus returned error: %v. Requeueing %s for tie-breaker.", err, task.TaskID)
		d.Queue.RequeueTask(task)
		return
	}

	// Emit per-page progress event so the frontend can update progress bars
	completedPages := d.Consensus.GetCompletedPageCount(task.JobID)
	if d.ctx != nil {
		percentage := (float64(completedPages) / float64(job.TotalTasks)) * 100
		runtime.EventsEmit(d.ctx, "job:progress", map[string]interface{}{
			"jobID":      task.JobID,
			"completed":  completedPages,
			"total":      job.TotalTasks,
			"percentage": percentage,
		})
	}

	// ok is true ONLY when this specific task pushed the final consensus over the finish line
	if ok {
		// Dedup guard: only emit document:complete once per job
		d.emitMu.Lock()
		if d.emittedJobs[task.JobID] {
			d.emitMu.Unlock()
			return
		}
		d.emittedJobs[task.JobID] = true
		d.emitMu.Unlock()

		fullText, _ := d.Consensus.GetVerifiedResult(task.JobID, job.TotalTasks)
		log.Printf("[Master] 🟢 JOB %s COMPLETED 100%%! emitting to Frontend.", task.JobID)

		// Emit event containing the final OCR text payload right to React
		if d.ctx != nil {
			runtime.EventsEmit(d.ctx, "document:complete", map[string]interface{}{
				"jobId":      task.JobID,
				"text":       fullText,
				"totalPages": job.TotalTasks,
			})
		}
	}
}

// GetCompletedJobs returns all fully verified jobs from the consensus engine.
func (d *Dispatcher) GetCompletedJobs() []CompletedJob {
	return d.Consensus.GetAllCompletedJobs(func(jobID string) int {
		job, exists := d.Queue.GetJob(jobID)
		if !exists {
			return 0
		}
		return job.TotalTasks
	})
}

// pickTwoDistinct randomly selects two distinct nodes from the list.
// Panics if len(nodes) < 2.
func pickTwoDistinct(nodes []swim.Node) (swim.Node, swim.Node) {
	n := len(nodes)
	if n < 2 {
		panic("Not enough nodes to pick two distinct ones")
	}

	idx1 := rand.Intn(n)
	idx2 := idx1
	for idx2 == idx1 {
		idx2 = rand.Intn(n)
	}

	return nodes[idx1], nodes[idx2]
}
