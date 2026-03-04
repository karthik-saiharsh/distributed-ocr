import { useState, useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetClusterNodes, ScanForNodes, UploadDocument } from '../wailsjs/go/main/App';
import { swim } from '../wailsjs/go/models';

const App = () => {
  const [nodes, setNodes] = useState<swim.Node[]>([]);
  const [scanning, setScanning] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<string>('');
  const [completedDocs, setCompletedDocs] = useState<{ jobId: string, text: string }[]>([]);

  useEffect(() => {
    // Populate the list immediately from whatever the backend already knows.
    GetClusterNodes().then((list: swim.Node[]) => setNodes(list ?? []));

    // Keep the list live: every time the SWIM layer discovers or loses a node
    // it emits "cluster:update" with the full current snapshot.
    const unsubCluster = EventsOn('cluster:update', (updated: swim.Node[]) => {
      setNodes(updated ?? []);
    });

    // Listen for Fully Completed OCR Job results from the Master Node
    const unsubDocs = EventsOn('document:complete', (payload: any) => {
      setCompletedDocs(prev => [{
        jobId: payload.jobId,
        text: payload.text
      }, ...prev]);
      setUploadStatus(`🎉 Document ${payload.jobId} fully extracted across the cluster!`);
    });

    // Clean up the listener if the component ever unmounts.
    return () => {
      unsubCluster();
      unsubDocs();
    };
  }, []);

  const handleScan = () => {
    setScanning(true);
    // ScanForNodes() returns immediately; the list updates reactively via
    // "cluster:update" events as peers respond to the subnet PINGs.
    ScanForNodes().finally(() => setScanning(false));
  };

  const handleUpload = () => {
    setUploadStatus('Dispatching tasks...');
    UploadDocument().then((numPages: number) => {
      setTimeout(() => setUploadStatus(`✅ ${numPages} Tasks successfully dispatched to cluster! Check terminal logs to watch the Workers steal them.`), 500);
      setTimeout(() => setUploadStatus(''), 7000);
    }).catch((err: any) => {
      setUploadStatus(`❌ Error: ${err}`);
      setTimeout(() => setUploadStatus(''), 5000);
    });
  };

  return (
    <div className="p-8 font-geist">
      <h1 className="text-2xl font-bold mb-6">Distributed OCR — Cluster</h1>

      <div className="flex gap-4 mb-6">
        <button
          onClick={handleScan}
          disabled={scanning}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed transition-colors"
        >
          {scanning ? 'Scanning…' : 'Scan for Nodes'}
        </button>

        <button
          onClick={handleUpload}
          className="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 cursor-pointer transition-colors flex items-center gap-2"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
          Upload Document
        </button>
      </div>

      {uploadStatus && (
        <div className="mb-6 p-4 rounded bg-gray-50 border border-gray-200 text-sm font-medium text-gray-800">
          {uploadStatus}
        </div>
      )}

      <div className="mt-6 border-t pt-4">
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

      {completedDocs.length > 0 && (
        <div className="mt-8 border-t pt-6">
          <h2 className="text-xl font-bold mb-4 text-green-700">✅ Distributed OCR Results</h2>
          <div className="flex flex-col gap-6">
            {completedDocs.map((doc, idx) => (
              <div key={idx} className="bg-white border text-left border-green-200 rounded p-4 shadow-sm">
                <p className="text-xs font-mono text-gray-500 mb-2 font-bold bg-gray-100 p-2 border-b block pb-2">Job Hash: {doc.jobId}</p>
                <div className="text-sm font-mono whitespace-pre-wrap p-2 text-gray-800">
                  {doc.text}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default App;
