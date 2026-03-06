# DOCR: Privacy-First Distributed OCR Grid

<p align="center">
  A local-first, distributed Optical Character Recognition (OCR) platform built with Go, React, and Wails.
</p>

## Overview

In the modern digital infrastructure, organizations face a massive bottleneck: the digitization of physical archives. Processing tens of thousands of high-resolution pages is computationally expensive and slow on a single machine. Cloud solutions (like AWS Textract) introduce severe data privacy concerns (HIPAA, GDPR) and require constant internet connectivity.

**Swarm** solves this by creating a **decentralized compute grid** out of ad-hoc local devices (laptops, desktops) sitting around in your office. It uses a **Master-Worker topology** over a Local Area Network (LAN) to securely and privately distribute OCR tasks using advanced distributed systems techniques.

---

## Key Features

*   **Privacy-First & Local:** Zero cloud dependency. Sensitive documents (medical records, legal contracts) never leave your local network. Air-gap friendly.
*   **Dynamic Load Balancing (Work Stealing):** Idle worker nodes proactively "steal" tasks from busy nodes via direct P2P RPC, ensuring maximum CPU utilization across the cluster.
*   **Autonomic Peer Discovery (SWIM Gossip):** Nodes dynamically form a mesh network via UDP multicast gossip. If a laptop is closed or disconnects, the cluster self-heals without task loss.
*   **Result Verification (Consensus):** Implements redundant execution. Multiple workers process the same chunk and the Master verifies consensus to defend against malicious nodes or corrupted processing.
*   **Cross-Platform GUI:** A sleek interface built with React, Vite, and Wails, giving a native desktop feel on Mac, Windows, and Linux.

---

## System Architecture

<img src="https://github.com/karthik-saiharsh/distributed-ocr/blob/main/Architechture.jpeg?raw=true">System Architechture</img>

The architecture relies on high-performance concurrent processing in Go and robust networking protocols:

1.  **Master Node (Orchestrator):** Runs the Wails GUI, manages the Global Job Queue, parses multi-page PDFs locally, and validates the integrity of returned OCR data.
2.  **Worker Nodes:** Stateless compute units running the Tesseract CGO wrapper. Features a local Double-Ended Queue (Deque) optimized for both LIFO local processing (cache locality) and FIFO work stealing.

*See [`explanation.md`](explanation.md) for a deep dive into the network topologies and data flow.*

---

## Prerequisites

To run or develop Swarm, ensure you have the following installed:

1.  **[Go](https://go.dev/doc/install)** (1.20+)
2.  **[Node.js & npm](https://nodejs.org/en/)**
3.  **[Wails Setup](https://wails.io/docs/gettingstarted/installation)**
4.  **Tesseract OCR:** Required on each machine for the core engine:
    *   **macOS:** `brew install tesseract`
    *   **Linux (Ubuntu):** `sudo apt-get install tesseract-ocr libtesseract-dev`
    *   **Windows:** UB-Mannheim Tesseract installer

*(Note: We use `github.com/gen2brain/go-fitz` for cross-platform PDF handling).*

---

## Development Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/your-org/distributed-ocr.git
    cd distributed-ocr
    ```

2.  **Frontend Setup:**
    The project uses a React/Vite frontend located in `/frontend`.
    ```bash
    cd frontend
    npm install
    npm run dev
    ```

3.  **Backend/App Setup:**
    The main Wails application is bound in `app.go`. To run the application in development mode with hot-reloading:
    ```bash
    # From the project root
    wails dev
    ```
    *Linux Users:* Run with `wails dev -tags webkit2_41` to support specific webkit dependencies.

---

## Running the Distributed Cluster

To see the distributed work stealing and gossip protocols in action across physical machines:

1.  **LAN Connection:** Connect multiple computers (e.g., Laptop A and Laptop B) to the exact same local Wi-Fi or router.
2.  **Build the Release:**
    Compile the app for production on both machines:
    ```bash
    wails build
    ```
3.  **Launch Nodes:** Open the compiled app executable (found in `build/bin/`) on both computers.
4.  **Discover Peers:** On Laptop A (your designated Master), click **Scan For Nodes**. The UDP gossip protocol will automatically discover Laptop B's IP address.
5.  **Distribute Work:** Click **Upload Document** on Laptop A and select a large PDF. 
6.  **Watch the Magic:** Laptop A chunks the PDF and distributes it via RPC. Laptop B will compute the OCR using its local Tesseract instance and return strings back to Laptop A to be verified and stitched back together!

---

## Repository Structure

*   `/frontend` - React, TypeScript, Vite frontend source.
*   `/master` - Orchestrator logic, consensus verification, and job queuing.
*   `/worker` - Node executor, Task Deque, and Tesseract C/Go bindings.
*   `/swim` - Custom UDP Gossip and discovery protocol implementation.
*   `/rpc` - Protobuf/TCP communication interfaces for task assignments and work stealing.

---

## Contributing

We welcome pull requests! 
1. Create a new branch for your feature (`git checkout -b feature/nice-feature`).
2. Make your backend changes in Go or frontend changes inside `/frontend`.
3. Please do **not** commit to `main` directly.
4. Submit a PR!

## License
GNU GPL V3
