# System Architecture: Swarm (Privacy-First OCR Grid)

## 0. Problem Statement
In the modern digital infrastructure, organizations such as hospitals, law firms, and archival institutions face a bottleneck: the digitization of massive physical archives. While Optical Character Recognition (OCR) technology exists to convert scanned documents into searchable text, processing tens of thousands of high-resolution pages is computationally expensive. A single modern workstation can take weeks to process a large archive, creating a significant backlog in data accessibility.

Traditionally, this computational load is offloaded to cloud-based services like AWS Textract or Google Cloud Vision. However, this approach introduces severe data security and privacy concerns, along with compliance risks. Transferring sensitive data—such as patient medical records or confidential legal contracts—to third-party cloud servers often violates strict privacy regulations like HIPAA (Health Insurance Portability and Accountability Act) and GDPR (General Data Protection Regulation). Furthermore, in environments with limited or intermittent internet connectivity (e.g., field hospitals, disaster zones, or secure air-gapped facilities), cloud dependency renders these services unusable.

Simultaneously, these organizations often have a large fleet of underutilized local hardware. Dozens of desktop computers and laptops that sit idle for the majority of the day.

This opens up scope for a local first distributed OCR platform.

---

## 1. High-Level Overview
The architecture follows a **Master-Worker topology** operating within a **Local Area Network (LAN)** boundary. The system is designed to process sensitive documents (e.g., medical records) using a decentralized compute grid formed by ad-hoc devices.

The architecture addresses three distributed systems challenges:
1.  **Dynamic Load Balancing:** Via the Work Stealing algorithm.
2.  **Fault Tolerance & Discovery:** Via the SWIM Gossip protocol.
3.  **Result Integrity:** Via Redundant Execution and Consensus verification.

---

## 2. Component Breakdown

### A. The Master Node (Orchestrator)
The Master Node serves as the entry point and coordination center for the cluster. It does not perform heavy computation (OCR) itself but manages the lifecycle of tasks.

* **Wails GUI (Frontend):**
    * The user interface built with React/Svelte and wrapped in Go using Wails.
    * **Responsibility:** Accepts PDF uploads from the user and visualizes cluster health (active nodes, progress bars).
* **Global Job Queue:**
    * **Responsibility:** Holds the backlog of pending pages (e.g., 10,000 pages from a scanned patient file).
    * **Mechanism:** Splits the input PDF into individual image chunks to be distributed.
* **Verification Logic (Consensus Check):**
    * **Responsibility:** Implements the "Integrity Upgrade." It receives results for the same Task ID from multiple workers and compares them.
    * **Logic:** `If Result(Worker A) == Result(Worker B) -> Accept`. If they mismatch, the task is re-queued for a tie-breaker.
* **Verified Results Store:**
    * **Responsibility:** Aggregates the final, trusted text data into a JSON or searchable PDF format for the user.

### B. The Worker Nodes (A, B, C)
These are the compute units (e.g., laptops of doctors/staff). They are stateless and identical in function.

* **Local Task Deque (Double-Ended Queue):**
    * **Structure:** A thread-safe data structure holding assigned image chunks.
    * **Behavior:** Operates as a **Stack (LIFO)** for local processing (to maximize cache locality) and a **Queue (FIFO)** for thieves (to steal the "coldest" tasks).
* **Tesseract OCR Engine (CGO Wrapper):**
    * **Responsibility:** The core processing unit. It uses a CGO wrapper (`gosseract`) to interface with the C++ Tesseract library, converting image byte streams into string text.

---

## 3. Protocol & Communication Layers

The diagram illustrates three distinct communication patterns, represented by different arrow styles:

### 1. SWIM Discovery Protocol (Dashed Lines)
* **Type:** UDP Multicast / Gossip
* **Purpose:** Failure Detection & Membership.
* **Mechanism:**
    * Nodes form a mesh network.
    * Every node periodically "gossips" with random peers (e.g., Worker A pings Worker B).
    * If a node stops responding (e.g., a laptop lid closes), the gossip propagates this "suspect" status to the Master, which removes the dead node from the active list.

### 2. Redundant Task Assignment (Solid Arrows)
* **Type:** TCP / RPC (Remote Procedure Call)
* **Purpose:** Fault Tolerance & Verification.
* **Mechanism:**
    * The diagram shows **Task ID: 101** being sent to **Worker A (Primary)**.
    * Simultaneously, the same **Task ID: 101** is sent to **Worker B (Redundant)**.
    * This ensures that even if Worker A is malicious or buggy, the error is caught by the Verification Logic when compared against Worker B.

### 3. Work Stealing (Curved Arrow)
* **Type:** RPC (Direct Peer-to-Peer)
* **Purpose:** Dynamic Load Balancing.
* **Mechanism:**
    * **Scenario:** Worker C has completed its tasks and is idle.
    * **Action (StealWorkRequest):** Worker C (the Thief) contacts Worker B (the Victim) directly, bypassing the Master.
    * **Response (Transfer Tasks):** Worker B moves half of its pending tasks from the *top* (head) of its deque and transfers them to Worker C.
    * **Benefit:** This prevents the Master from becoming a bottleneck and ensures all CPU resources remain utilized.

---

## 4. Operational Workflow (Data Flow)

1.  **Ingestion:** User drags a PDF into the **Wails GUI** on the Master Node.
2.  **Queuing:** The Master splits the PDF and pushes chunks to the **Global Job Queue**.
3.  **Assignment:** The Master pushes `Page_1.jpg` to **Worker A** and **Worker B** via RPC.
4.  **Processing:** Both workers process the image using **Tesseract OCR**.
5.  **Consensus:** Both workers return the text "Diagnosis: Fever" to the Master.
6.  **Verification:** The **Verification Logic** confirms the match and saves it to the **Results Store**.
7.  **Rebalancing:** Meanwhile, **Worker C** finishes early, detects **Worker B** is busy, and performs a **Work Steal** operation to help finish the remaining queue.
