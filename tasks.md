Task List: SWIM Discovery Implementation
Phase 1: Backend Data Structures (Go)

    Define the Node Model: Create a struct to represent a single machine. It must include fields for Unique ID, IP Address, Port, Status (Alive/Suspect/Dead), and a "Last Updated" timestamp.

    Define the MembershipList: Create a container to hold all known nodes. This must be a Map (for O(1) lookups) guarded by a Read/Write Mutex to prevent race conditions between the UDP listener and the UI.

    Define the Protocol Message Format: Create a struct for the data sent over the network. It needs fields for MessageType (PING, ACK, PING-REQ), SenderID, and a Payload (to carry the gossip list).

Phase 2: Networking Layer (Go)

    Initialize UDP Listener: specific logic to open a UDP socket on a configurable port when the application starts.

    Implement Packet Receiver: Create a continuous loop that reads incoming bytes from the UDP socket and unmarshals the JSON data into your Message struct.

    Implement Message Router: Write logic to switch based on MessageType.

        If PING: Update the sender's timestamp in the local list and immediately reply.

        If ACK: Mark the sender as "Alive" and cancel any pending timeout timers.

Phase 3: Gossip & Failure Detection Logic (Go)

    Implement the Heartbeat Ticker: Create a background routine that triggers at a fixed interval (e.g., every 500ms).

    Implement Random Peer Selection: On every tick, write logic to select one random node from the MembershipList (excluding self).

    Implement the PING Sender: Send a PING message to the selected peer and start a short timeout timer (e.g., 200ms).

    Implement Timeout Handler: If the timer expires before an ACK is received, update that node's status to SUSPECT.

    Implement State Merging (Gossip): When receiving a PING or ACK, read the attached list of members. Compare it with your local list and add any new nodes found (Discovery).

Phase 4: Wails Bridge (Go to JS)

    Expose Initial State: Create a public Wails method (e.g., GetClusterNodes) that returns the current MembershipList so the frontend can render immediately on load.

    Implement Event Emitter: In your networking logic, whenever a node is added, removed, or changes status, trigger a Wails Event (e.g., cluster:update) with the new list as the payload.

Phase 5: Frontend Visualization (React)

    Create Cluster Context/Store: Set up a React state variable (e.g., useState or a Context) to hold the array of nodes.

    Implement Event Listener: Use the Wails runtime to subscribe to the cluster:update event. Update your React state whenever this event fires.

    Build the Node Grid Component: Create a UI component that maps through the node array.

        Visual Feedback: Style the cards dynamically based on status (e.g., Green border for Alive, Red background for Dead/Suspect).

        Info Display: Show the Node ID and IP address on the card.

    Testing Dashboard: (Optional) Add a button to the UI that calls a Go method to manually "Kill" the local node (stop the heartbeat) to test how other nodes react.