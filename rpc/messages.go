package rpc

// TaskRequest represents a chunk of work sent from Master to Worker.
type TaskRequest struct {
	TaskID    string `json:"taskId"` // Unique ID for this specific work item (e.g., DocID_PageNum)
	JobID     string `json:"jobId"`  // ID of the overarching document being processed
	ImageData []byte `json:"-"`      // The actual image slice to process (omitted from JSON logs usually)
	PageNum   int    `json:"pageNum"`
}

// TaskResponse represents the result of a processed chunk sent back to Master.
type TaskResponse struct {
	TaskID        string `json:"taskId"`
	WorkerID      string `json:"workerId"`
	ExtractedText string `json:"extractedText"`
	Error         string `json:"error,omitempty"`
}

// StealRequest represents a peer-to-peer request for work.
type StealRequest struct {
	ThiefID string `json:"thiefId"`
	Count   int    `json:"count"` // How many tasks the thief wants to steal
}

// StealResponse contains the stolen tasks.
type StealResponse struct {
	Tasks []TaskRequest `json:"tasks"`
}
