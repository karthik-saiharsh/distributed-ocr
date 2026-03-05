import { useState, useEffect, useCallback } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetClusterNodes, ScanForNodes, UploadDocument, KillNode, GetQueueDepth, GetCompletedJobs } from '../wailsjs/go/main/App';
import { swim, master } from '../wailsjs/go/models';

import Sidebar, { type Page } from './components/Sidebar';
import Dashboard from './pages/Dashboard';
import Cluster from './pages/Cluster';
import Jobs from './pages/Jobs';
import Results from './pages/Results';
import Settings from './pages/Settings';

export interface JobInfo {
  id: string;
  name: string;
  totalPages: number;
  completedPages: number;
  status: 'queued' | 'processing' | 'completed';
}

const App = () => {
  const [activePage, setActivePage] = useState<Page>('dashboard');
  const [nodes, setNodes] = useState<swim.Node[]>([]);
  const [scanning, setScanning] = useState(false);
  const [completedDocs, setCompletedDocs] = useState<{ jobId: string; text: string; totalPages: number }[]>([]);
  const [jobs, setJobs] = useState<JobInfo[]>([]);
  const [queueDepth, setQueueDepth] = useState(0);

  // Poll queue depth periodically
  useEffect(() => {
    const interval = setInterval(() => {
      GetQueueDepth().then((d: number) => setQueueDepth(d ?? 0)).catch(() => { });
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    // Populate from backend on mount.
    GetClusterNodes().then((list: swim.Node[]) => setNodes(list ?? []));
    // Load any already completed jobs
    GetCompletedJobs().then((completed: master.CompletedJob[] | null) => {
      if (completed && completed.length > 0) {
        setCompletedDocs(completed.map((j) => ({
          jobId: j.jobId,
          text: j.text,
          totalPages: j.totalPages,
        })));
      }
    }).catch(() => { });

    // Subscribe to live membership changes.
    const unsub1 = EventsOn('cluster:update', (updated: swim.Node[]) => {
      setNodes(updated ?? []);
    });

    // Subscribe to job progress events.
    const unsub2 = EventsOn('job:progress', (data: any) => {
      if (!data) return;
      setJobs((prev) => {
        const idx = prev.findIndex((j) => j.id === data.jobId);
        if (idx === -1) return prev; // Job not tracked yet
        const updated = [...prev];
        updated[idx] = {
          ...updated[idx],
          completedPages: data.completedPages ?? updated[idx].completedPages,
          totalPages: data.totalPages ?? updated[idx].totalPages,
          status: (data.completedPages >= data.totalPages) ? 'completed' : 'processing',
        };
        return updated;
      });
    });

    // Subscribe to document completion events.
    const unsub3 = EventsOn('document:complete', (data: any) => {
      if (!data) return;
      setCompletedDocs((prev) => {
        // Dedup: don't add the same job twice
        if (prev.some((d) => d.jobId === data.jobId)) return prev;
        return [...prev, {
          jobId: data.jobId,
          text: data.text ?? '',
          totalPages: data.totalPages ?? 0,
        }];
      });
      // Mark job as completed
      setJobs((prev) =>
        prev.map((j) =>
          j.id === data.jobId
            ? { ...j, status: 'completed' as const, completedPages: j.totalPages }
            : j
        )
      );
    });

    return () => {
      unsub1();
      unsub2();
      unsub3();
    };
  }, []);

  const handleScan = () => {
    setScanning(true);
    ScanForNodes().finally(() => setScanning(false));
  };

  const uploadCount = jobs.length;

  const handleUpload = useCallback(() => {
    UploadDocument().then((numPages: number) => {
      const jobId = `job-${Date.now()}`;
      const newJob: JobInfo = {
        id: jobId,
        name: `Document ${uploadCount + 1}`,
        totalPages: numPages,
        completedPages: 0,
        status: 'queued',
      };
      setJobs((prev) => [newJob, ...prev]);

      // After a brief delay, mark as processing (tasks are being dispatched async)
      setTimeout(() => {
        setJobs((prev) =>
          prev.map((j) => (j.id === jobId ? { ...j, status: 'processing' as const } : j))
        );
      }, 1500);
    }).catch(() => { });
  }, [uploadCount]);

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
            completedCount={completedDocs.length}
            queueDepth={queueDepth}
            jobCount={jobs.length}
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
        return <Jobs jobs={jobs} onUpload={handleUpload} />;
      case 'results':
        return <Results completedDocs={completedDocs} />;
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
