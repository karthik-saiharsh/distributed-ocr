import { useState, useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetClusterNodes, ScanForNodes } from '../wailsjs/go/main/App';
import { swim } from '../wailsjs/go/models';

const App = () => {
  const [nodes, setNodes] = useState<swim.Node[]>([]);
  const [scanning, setScanning] = useState(false);

  useEffect(() => {
    // Populate the list immediately from whatever the backend already knows.
    GetClusterNodes().then((list: swim.Node[]) => setNodes(list ?? []));

    // Keep the list live: every time the SWIM layer discovers or loses a node
    // it emits "cluster:update" with the full current snapshot.
    const unsub = EventsOn('cluster:update', (updated: swim.Node[]) => {
      setNodes(updated ?? []);
    });

    // Clean up the listener if the component ever unmounts.
    return () => { unsub(); };
  }, []);

  const handleScan = () => {
    setScanning(true);
    // ScanForNodes() returns immediately; the list updates reactively via
    // "cluster:update" events as peers respond to the subnet PINGs.
    ScanForNodes().finally(() => setScanning(false));
  };

  return (
    <div className="p-8 font-geist">
      <h1 className="text-2xl font-bold mb-6">Distributed OCR — Cluster</h1>

      <button
        onClick={handleScan}
        disabled={scanning}
        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
      >
        {scanning ? 'Scanning…' : 'Scan for Nodes'}
      </button>

      <div className="mt-6">
        <p className="font-semibold mb-2">Nodes found: {nodes.length}</p>
        {nodes.length === 0 ? (
          <p className="text-gray-500">No nodes discovered yet. Press "Scan for Nodes" to start.</p>
        ) : (
          nodes.map((node: swim.Node) => (
            <p key={node.id} className="font-mono text-sm mt-1">
              [{node.status}] {node.ip}:{node.port} — {node.id}
            </p>
          ))
        )}
      </div>
    </div>
  );
};

export default App;
