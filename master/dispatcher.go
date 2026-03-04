package master

import (
	"dist-ocr/rpc"
	"dist-ocr/swim"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Dispatcher constantly monitors the GlobalQueue and pushes tasks to workers via RPC.
type Dispatcher struct {
	Queue     *GlobalQueue
	Consensus *ConsensusEngine
	cluster   *swim.SWIMService
	stopCh    chan struct{}
	
	// RPC Port for connecting to workers
	rpcPort int
}

func NewDispatcher(cluster *swim.SWIMService, rpcPort int) *Dispatcher {
	return &Dispatcher{
		Queue:     NewGlobalQueue(),
		Consensus: NewConsensusEngine(),
		cluster:   cluster,
		stopCh:    make(chan struct{}),
		rpcPort:   rpcPort,
	}
}

// Start begins the background dispatch loop.
func (d *Dispatcher) Start() {
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
			for _, n := range allNodes {
				if n.Status == swim.StatusAlive {
					aliveNodes = append(aliveNodes, n)
				}
			}

			// We need at least 2 alive nodes to do redundant assignment!
			// (During testing you might allow 1, but MVP architecture demands 2)
			if len(aliveNodes) < 2 {
				continue
			}

			// Grab a task
			task, ok := d.Queue.PopTask()
			if !ok {
				continue
			}

			// Pick two distinct random workers to assign this same task to
			w1, w2 := pickTwoDistinct(aliveNodes)

			// Dispatch asynchronously
			go d.assignAndVerify(task, w1)
			go d.assignAndVerify(task, w2)
		}
	}
}

// assignAndVerify connects to a worker via RPC, sends the task, waits for response, and submits to Consensus.
func (d *Dispatcher) assignAndVerify(task rpc.TaskRequest, worker swim.Node) {
	addr := fmt.Sprintf("%s:%d", worker.IP, d.rpcPort) // Connect to the Worker's RPC port
	
	client, err := rpc.NewClient(addr)
	if err != nil {
		log.Printf("[Master] Failed to dial worker %s: %v. Requeueing task %s", worker.ID, err, task.TaskID)
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
	_, err = d.Consensus.SubmitResult(*resp, task.PageNum, task.JobID)
	if err != nil {
		// Mismatch! The consensus engine tells us to requeue
		log.Printf("[Master] Consensus returned error: %v. Requeueing %s for tie-breaker.", err, task.TaskID)
		d.Queue.RequeueTask(task)
		return
	}
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
