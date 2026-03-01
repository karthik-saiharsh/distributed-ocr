package main

import (
	"context"
	"fmt"

	"dist-ocr/swim"

	"github.com/google/uuid"
)

// App struct holds the Wails application context and the SWIM service.
type App struct {
	ctx         context.Context
	swimService *swim.SWIMService
}

// create a new App application struct.
func NewApp() *App {
	// Detect the local LAN IP automatically.
	localIP := swim.DetectLocalIP()

	// Generate a stable node ID for this instance.
	// uuid is universally unique identifier.
	// each node gets a uuid on startup as an identification number
	nodeID := uuid.New().String()

	svc := swim.NewSWIMService(nodeID, localIP, 7946)
	return &App{swimService: svc}
}

// startup is called when the app starts. The context is saved so we can call
// Wails runtime methods (e.g. EventsEmit). The SWIM service is started here.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.swimService.Start(ctx); err != nil {
		fmt.Printf("[App] Failed to start SWIM service: %v\n", err)
	}
}

// shutdown is called when the application is about to quit.
func (a *App) shutdown(_ context.Context) {
	a.swimService.Stop()
}

// GetClusterNodes is a Wails bound method that returns the current cluster
// membership list. The frontend calls this on initial load to populate the UI.
func (a *App) GetClusterNodes() []swim.Node {
	return a.swimService.GetClusterNodes()
}

// ScanForNodes is a Wails bound method called by the frontend.
// It fires a PING to every address on the local subnet.
// all peers currently running the app are discovered. The call returns
// immediately; membership updates arrive asynchronously via "cluster:update"
// events as peers reply with ACKs.
func (a *App) ScanForNodes() {
	go a.swimService.StartScan()
}

// KillNode stops the local SWIM heartbeat, basically a node failure.
func (a *App) KillNode() {
	a.swimService.KillNode()
}
