package master

import (
	"dist-ocr/rpc"
	"fmt"
	"log"
	"sync"
)

// ConsensusEngine is responsible for gathering redundant results from workers
// and verifying that they match before accepting the text.
type ConsensusEngine struct {
	mu sync.Mutex
	// Maps TaskID -> Slice of Responses from different workers
	results map[string][]rpc.TaskResponse
	// Maps JobID -> Verified fully aggregated text for that job
	verifiedData map[string]map[int]string // JobID -> PageNum -> Text
}

func NewConsensusEngine() *ConsensusEngine {
	return &ConsensusEngine{
		results:      make(map[string][]rpc.TaskResponse),
		verifiedData: make(map[string]map[int]string),
	}
}

// SubmitResult is called by the Dispatcher when a worker returns a response over RPC.
// Returns (isJobComplete, error)
func (c *ConsensusEngine) SubmitResult(resp rpc.TaskResponse, pageNum int, jobID string, expectedPages int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we already have verified this task (e.g. from a previous successful round)
	if _, ok := c.verifiedData[jobID]; !ok {
		c.verifiedData[jobID] = make(map[int]string)
	}
	if _, verified := c.verifiedData[jobID][pageNum]; verified {
		return true, nil // Already verified, ignore this late submission
	}

	// 1. Add this worker's result
	c.results[resp.TaskID] = append(c.results[resp.TaskID], resp)
	submissions := c.results[resp.TaskID]

	log.Printf("[Master] Received result for Task %s from Worker %s. Total subs: %d", resp.TaskID, resp.WorkerID, len(submissions))

	// 2. Wait until we have at least 2 results to compare
	if len(submissions) < 2 {
		return false, nil
	}

	// 3. We have 2+ results. We implement the "Integrity Upgrade"
	// For MVP, we literally just compare Result A == Result B.

	resultA := submissions[0].ExtractedText
	resultB := submissions[1].ExtractedText

	if resultA == resultB {
		log.Printf("[Master] Consensus REACHED for Task %s!", resp.TaskID)
		c.verifiedData[jobID][pageNum] = resultA

		// Clean up the pending results map for memory
		delete(c.results, resp.TaskID)

		// Check if the entire job is now fully complete!
		if len(c.verifiedData[jobID]) == expectedPages {
			return true, nil // Signal the UI that this job is finished
		}

		return false, nil
	}

	// 4. Mismatch!
	// This worker gave a different string than the previous worker.
	log.Printf("[Master] Consensus MISMATCH for Task %s! Req-queueing for tie-breaker.", resp.TaskID)

	// We delete the results so it starts fresh when re-queued.
	delete(c.results, resp.TaskID)

	return false, fmt.Errorf("consensus mismatch between workers")
}

// GetVerifiedResult returns the combined text for a complete job, if all pages are done.
func (c *ConsensusEngine) GetVerifiedResult(jobID string, expectedPages int) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pages, ok := c.verifiedData[jobID]
	if !ok {
		return "", false
	}

	if len(pages) < expectedPages {
		return "", false // Not finished yet
	}

	// Stitch them together in order
	var fullText string
	for i := 1; i <= expectedPages; i++ {
		fullText += pages[i] + "\n\n"
	}

	return fullText, true
}

// GetCompletedPageCount returns how many pages have been verified for a job.
func (c *ConsensusEngine) GetCompletedPageCount(jobID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	pages, ok := c.verifiedData[jobID]
	if !ok {
		return 0
	}
	return len(pages)
}

// CompletedJob holds the metadata for a finished job.
type CompletedJob struct {
	JobID      string `json:"jobId"`
	Text       string `json:"text"`
	TotalPages int    `json:"totalPages"`
}

// GetAllCompletedJobs returns a list of all jobs that have been fully verified.
// It requires a lookup function to resolve totalPages for each job.
func (c *ConsensusEngine) GetAllCompletedJobs(getJobTotal func(jobID string) int) []CompletedJob {
	c.mu.Lock()
	defer c.mu.Unlock()

	var completed []CompletedJob
	for jobID, pages := range c.verifiedData {
		total := getJobTotal(jobID)
		if total > 0 && len(pages) >= total {
			var fullText string
			for i := 1; i <= total; i++ {
				fullText += pages[i] + "\n\n"
			}
			completed = append(completed, CompletedJob{
				JobID:      jobID,
				Text:       fullText,
				TotalPages: total,
			})
		}
	}
	return completed
}
