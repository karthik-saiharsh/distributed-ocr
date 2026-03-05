import { useState, useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetClusterNodes, ScanForNodes, UploadDocument, KillNode } from '../wailsjs/go/main/App';
import { swim } from '../wailsjs/go/models';

import Sidebar, { type Page } from './components/Sidebar';
import Dashboard from './pages/Dashboard';
import Cluster from './pages/Cluster';
import Jobs from './pages/Jobs';
import Results from './pages/Results';
import Settings from './pages/Settings';

const App = () => {
  const [activePage, setActivePage] = useState<Page>('dashboard');
  const [nodes, setNodes] = useState<swim.Node[]>([]);
  const [scanning, setScanning] = useState(false);

  useEffect(() => {
    // Populate from backend on mount.
    GetClusterNodes().then((list: swim.Node[]) => setNodes(list ?? []));

    // Subscribe to live membership changes.
    const unsub = EventsOn('cluster:update', (updated: swim.Node[]) => {
      setNodes(updated ?? []);
    });

    return () => { unsub(); };
  }, []);

  const handleScan = () => {
    setScanning(true);
    ScanForNodes().finally(() => setScanning(false));
  };

  const handleUpload = () => {
    UploadDocument();
  };

  const handleKillNode = () => {
    KillNode();
  };

  // Find self node (first node is usually self, as SWIM registers it on start).
  const selfNode = nodes.length > 0 ? nodes[0] : undefined;

  const renderPage = () => {
    switch (activePage) {
      case 'dashboard':
        return (
          <Dashboard
            nodes={nodes}
            selfNodeId={selfNode?.id}
            scanning={scanning}
            onScan={handleScan}
            onUpload={handleUpload}
            onNavigate={(p) => setActivePage(p as Page)}
          />
        );
      case 'cluster':
        return (
          <Cluster
            nodes={nodes}
            selfNodeId={selfNode?.id}
            scanning={scanning}
            onScan={handleScan}
          />
        );
      case 'jobs':
        return <Jobs onUpload={handleUpload} />;
      case 'results':
        return <Results />;
      case 'settings':
        return <Settings selfNode={selfNode} onKillNode={handleKillNode} />;
      default:
        return null;
    }
  };

  return (
    <div className="app-layout">
      <Sidebar
        activePage={activePage}
        onNavigate={setActivePage}
        selfNodeIP={selfNode?.ip}
        selfNodePort={selfNode?.port}
      />
      <main className="main-content">
        {renderPage()}
      </main>
    </div>
  );
};

export default App;
