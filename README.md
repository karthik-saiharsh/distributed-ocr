# README

## Project Setup

#### Step 1: Installing Wails:
- To Install Wails, you must have Go installed first. (https://go.dev/doc/install)
- After you install Go, to install visit (https://wails.io/docs/gettingstarted/installation) instructions are available for windows, mac and linux

> Note: A few additional steps are required for setting up wails on linux, this is also mentioned in the installation link.

#### Step 2: Development
- The Project has a `/frontend` directory which has the source for all the frontend, with `react` and `vite` project with `npm`.
- For developing frontend, `cd` into `/frontend` and run `npm run dev` (you have to install dependencies first (`npm install`))
- For those developing the backend, the `app.go` is the file to expose all the functions that will be available in the frontend.
- To run the backend you have to run `wails dev` in the root of this project.
- On linux alone you have to run it with `wails dev -tags webkit2_41`.
- `main.go` has the setup and details of the App. There mostly won't be any need to edit this file.
- Lastly, any new feature you plan to add, make a new branch and send a PR, do not commit to main directly.
- Anytime you want to change the frontend, install new dependencies or tools, or run any `npm` related command, do it only in `/frontend` directory.

## Distributed OCR (Work Stealing & Clustering)

This project features a fully distributed architecture using SWIM gossip and RPC to asynchronously steal OCR tasks from busy nodes and process physical PDF image bytes across the LAN.

### Dependencies
To utilize the distributed engine, each machine must have Tesseract OCR installed locally:
- **macOS:** `brew install tesseract`
- **Linux (Ubuntu):** `sudo apt-get install tesseract-ocr`
- **Windows:** Download the installer from the UB-Mannheim Tesseract project.

*(Note: `github.com/gen2brain/go-fitz` is used natively to process PDFs cross-platform without needing `poppler` bindings.)*

### Running the Cluster Across Multiple Computers
To run the distributed work stealing architecture legitimately across separate physical machines:

1. **Connect to LAN:** Ensure Laptop A and Laptop B are connected to the exact same WiFi connection or local network router.
2. **Compile the App:** Clone the repository onto both machines and compile the production application:
   ```bash
   wails build
   ```
3. **Launch Nodes:** Open the compiled application executable (`./build/bin/dist-ocr.app/Contents/MacOS/dist-ocr`) on both laptops. You do not need to specify any special ports.
4. **Discover Peers:** On Laptop A, click **Scan For Nodes** to initiate the UDP peer-discovery broadcast. You will see Laptop B's real LAN IP address appear in your list!
5. **Start Distributed Processing:** Click **Upload Document** on Laptop A and select a multi-page PDF.
6. **Watch the Magic:** Watch the graphical React UI and terminal logs. Laptop A (the Master Dispatcher) will chunk the PDF into physical image bytes and instantly distribute them via RPC over the WiFi network. Laptop B will securely steal the tasks, compute Tesseract OCR algorithms on its own hardware, and return the extracted strings back to Laptop A, where Consensus is verified and the full document is displayed in the UI!