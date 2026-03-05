package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"dist-ocr/master"
	"dist-ocr/rpc"
	"dist-ocr/swim"
	"dist-ocr/worker"

	"github.com/gen2brain/go-fitz"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the Wails application context and the SWIM service.
type App struct {
	ctx         context.Context
	swimService *swim.SWIMService

	// New MVP Components
	rpcServer  *rpc.Server
	executor   *worker.Executor
	dispatcher *master.Dispatcher
}

// create a new App application struct.
func NewApp() *App {
	// Detect the local LAN IP automatically.
	localIP := swim.DetectLocalIP()

	// Parse an optional base port flag (defaults to 7946)
	// This ensures we can run multiple instances locally without "address already in use" errors.
	var basePort int
	flag.IntVar(&basePort, "port", 7946, "Base port for SWIM and RPC")
	if !flag.Parsed() {
		flag.Parse()
	}

	nodeID := uuid.New().String()
	svc := swim.NewSWIMService(nodeID, localIP, basePort)

	// Run the RPC server on port+1 (e.g. 7947 by default)
	rpcPort := basePort + 1
	rpcServ := rpc.NewServer(rpcPort)

	// Create Worker Engine
	exec := worker.NewExecutor(nodeID, svc, rpcPort)

	// Create Master Engine
	disp := master.NewDispatcher(svc, rpcPort)

	return &App{
		swimService: svc,
		rpcServer:   rpcServ,
		executor:    exec,
		dispatcher:  disp,
	}
}

// startup is called when the app starts. The context is saved so we can call
// Wails runtime methods (e.g. EventsEmit). The SWIM service is started here.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Start SWIM Discovery
	if err := a.swimService.Start(ctx); err != nil {
		fmt.Printf("[App] Failed to start SWIM service: %v\n", err)
	}

	// 2. Start Worker RPC Server and Background Stealing Loop
	workerRPC := worker.NewWorkerRPC(a.executor)
	if err := a.rpcServer.Start(workerRPC); err != nil {
		fmt.Printf("[App] Failed to start RPC Server: %v\n", err)
	}
	a.executor.StartBackgroundWorker()

	// 3. Start Master Dispatcher
	a.dispatcher.Start(a.ctx)
}

// shutdown is called when the application is about to quit.
func (a *App) shutdown(_ context.Context) {
	a.dispatcher.Stop()
	a.rpcServer.Stop()
	a.swimService.Stop()
}

// GetClusterNodes is a Wails bound method that returns the current cluster
// membership list. The frontend calls this on initial load to populate the UI.
func (a *App) GetClusterNodes() []swim.Node {
	return a.swimService.GetClusterNodes()
}

// ScanForNodes is a Wails bound method called by the frontend.
func (a *App) ScanForNodes() {
	go a.swimService.StartScan()
}

// KillNode stops the local SWIM heartbeat, basically a node failure.
func (a *App) KillNode() {
	a.swimService.KillNode()
}

// UploadDocument generates a job with tasks to test the OCR pipeline.
// Exposed to Wails frontend. It returns the total number of pages/tasks that will be queued.
func (a *App) UploadDocument() (int, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select PDF Document to Distribute",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files", Pattern: "*.pdf"},
		},
	})
	if err != nil || selection == "" {
		log.Printf("[App] Document upload cancelled or failed: %v", err)
		return 0, fmt.Errorf("upload cancelled")
	}

	jobID := uuid.New().String()
	log.Printf("[App] Processing real document upload %s for Job ID: %s", selection, jobID)

	// Quickly evaluate the PDF dimensions synchronously so we can return the task count to React
	doc, err := fitz.New(selection)
	if err != nil {
		log.Printf("[App] Failed to open PDF: %v", err)
		return 0, err
	}
	numPages := doc.NumPage()

	// 2. Process image extraction asynchronously so the UI does not freeze!
	go func() {
		defer doc.Close()

		log.Printf("[App] PDF loaded cleanly. Slicing %d pages into images...", numPages)

		job := &master.Job{
			ID:         jobID,
			TotalTasks: numPages,
			Tasks:      make([]rpc.TaskRequest, numPages),
		}

		for i := 0; i < numPages; i++ {
			// Extract page as PNG at 300 DPI for high-quality OCR accuracy
			log.Printf("[App] Extracting Page %d/%d to internal memory...", i+1, numPages)
			imgBytes, err := doc.ImagePNG(i, 300.0)
			if err != nil {
				log.Printf("[App] Failed to extract page %d: %v", i+1, err)
				continue
			}

			job.Tasks[i] = rpc.TaskRequest{
				TaskID:    fmt.Sprintf("%s_Pg%d", jobID, i+1),
				JobID:     jobID,
				PageNum:   i + 1,
				ImageData: imgBytes, // Ship physical bytes over RPC
			}
		}

		// Dump them into the active Master Queue
		log.Printf("[App] PDF Extraction Complete! %d Tasks queued for distribution.", numPages)
		a.dispatcher.Queue.AddJob(job)
	}()

	return numPages, nil
}

// GetQueueDepth is a Wails bound method that returns the current number of pending tasks.
func (a *App) GetQueueDepth() int {
	return a.dispatcher.Queue.Len()
}

// GetCompletedJobs is a Wails bound method that returns all completed and verified OCR results.
func (a *App) GetCompletedJobs() []master.CompletedJob {
	return a.dispatcher.GetCompletedJobs()
}
