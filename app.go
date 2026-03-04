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

	"github.com/google/uuid"
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

	// Create Worker Engine
	exec := worker.NewExecutor(nodeID)

	// Run the RPC server on port+1 (e.g. 7947 by default)
	rpcPort := basePort + 1
	rpcServ := rpc.NewServer(rpcPort)

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

	// 2. Start Worker RPC Server
	workerRPC := worker.NewWorkerRPC(a.executor)
	if err := a.rpcServer.Start(workerRPC); err != nil {
		fmt.Printf("[App] Failed to start RPC Server: %v\n", err)
	}

	// 3. Start Master Dispatcher
	a.dispatcher.Start()
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
// Exposed to Wails frontend.
func (a *App) UploadDocument() {
	jobID := uuid.New().String()
	log.Printf("[App] Processing document upload for Job ID: %s", jobID)

	job := &master.Job{
		ID:         jobID,
		TotalTasks: 5, // A 5 page document
		Tasks:      make([]rpc.TaskRequest, 5),
	}

	for i := 0; i < 5; i++ {
		job.Tasks[i] = rpc.TaskRequest{
			TaskID:  fmt.Sprintf("%s_Pg%d", jobID, i+1),
			JobID:   jobID,
			PageNum: i + 1,
		}
	}

	// Dump them into the active Master Queue
	a.dispatcher.Queue.AddJob(job)
}
