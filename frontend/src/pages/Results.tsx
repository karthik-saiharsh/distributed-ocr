import React, { useState, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { IconResults, IconFileText, IconShield, IconCheck, IconX, IconActivity } from '../components/Icons';

interface ResultsProps {
    completedDocs: { jobId: string; text: string; totalPages: number }[];
}

interface NodeStats {
    [nodeId: string]: number;
}

const Results: React.FC<ResultsProps> = ({ completedDocs }) => {
    const [expandedJobId, setExpandedJobId] = useState<string | null>(null);
    const [copySuccess, setCopySuccess] = useState<string | null>(null);

    // Progress State
    const [progress, setProgress] = useState(0);
    const [completed, setCompleted] = useState(0);
    const [total, setTotal] = useState(0);
    const [nodeStats, setNodeStats] = useState<NodeStats>({});

    useEffect(() => {
        const unsub = EventsOn("ocr:progress", (data: { completed: number; percentage: number; total: number; nodeID: string; }) => {
            if (!data) return;
            
            // If it's a new document (completed count reset or total changed significantly and we're at 1)
            if (data.completed === 1 && data.percentage < 5) {
                setNodeStats({});
            }

            setProgress(data.percentage);
            setCompleted(data.completed);
            setTotal(data.total);

            if (data.nodeID) {
                setNodeStats(prev => ({
                    ...prev,
                    [data.nodeID]: (prev[data.nodeID] || 0) + 1
                }));
            }
        });

        return () => {
            unsub();
        };
    }, []);

    const handleCopy = (text: string, jobId: string) => {
        navigator.clipboard.writeText(text).then(() => {
            setCopySuccess(jobId);
            setTimeout(() => setCopySuccess(null), 2000);
        }).catch(() => { });
    };

    const toggleExpand = (jobId: string) => {
        setExpandedJobId((prev) => (prev === jobId ? null : jobId));
    };

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Results</h1>
                    <p className="page-subtitle">View verified OCR outputs from processed documents</p>
                </div>
            </div>

            {completedDocs.length === 0 ? (
                <>
                    <div className="card">
                        <div className="empty-state">
                            <div className="empty-state-icon">
                                <IconResults width={28} height={28} />
                            </div>
                            <div className="empty-state-title">No verified results yet</div>
                            <div className="empty-state-desc">
                                Results will appear here after documents are processed and consensus is reached between worker nodes.
                            </div>
                        </div>
                    </div>

                    {/* How It Works */}
                    <div className="section-gap">
                        <h3 style={{ fontSize: '16px', fontWeight: 600, marginBottom: '16px' }}>How Verification Works</h3>
                        <div className="grid-2">
                            <div className="card">
                                <div className="card-body" style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                                    <div
                                        className="list-item-icon"
                                        style={{
                                            background: 'rgba(99, 102, 241, 0.12)',
                                            color: 'var(--accent-blue)',
                                            flexShrink: 0,
                                        }}
                                    >
                                        <IconFileText width={20} height={20} />
                                    </div>
                                    <div>
                                        <div style={{ fontWeight: 600, fontSize: '14px', marginBottom: '4px' }}>Redundant Processing</div>
                                        <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                                            Each page is sent to at least two worker nodes for independent OCR extraction.
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className="card">
                                <div className="card-body" style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                                    <div
                                        className="list-item-icon"
                                        style={{
                                            background: 'var(--accent-green-dim)',
                                            color: 'var(--accent-green)',
                                            flexShrink: 0,
                                        }}
                                    >
                                        <IconShield width={20} height={20} />
                                    </div>
                                    <div>
                                        <div style={{ fontWeight: 600, fontSize: '14px', marginBottom: '4px' }}>Consensus Check</div>
                                        <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                                            Results are compared — if they match, the text is accepted. Mismatches trigger a tie-breaker round.
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </>
            ) : null}

            {/* Active Progress Tracker */}
            {total > 0 && progress < 100 && (
                <div className="card" style={{ marginBottom: '24px', border: '1px solid var(--accent-blue)' }}>
                    <div className="card-header" style={{ paddingBottom: '12px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <div className="list-item-icon" style={{ background: 'rgba(99, 102, 241, 0.12)', color: 'var(--accent-blue)' }}>
                                <IconActivity width={18} height={18} className="animate-pulse" />
                            </div>
                            <div>
                                <div className="card-title">Processing Document...</div>
                                <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                                    {completed} / {total} pages verified
                                </div>
                            </div>
                        </div>
                        <div style={{ fontWeight: 600, color: 'var(--accent-blue)', fontSize: '18px' }}>
                            {Math.round(progress)}%
                        </div>
                    </div>
                    
                    <div className="card-body" style={{ paddingTop: 0 }}>
                        <div className="progress-bar-track" style={{ height: '8px', marginBottom: '16px' }}>
                            <div 
                                className="progress-bar-fill" 
                                style={{ 
                                    width: `${progress}%`, 
                                    background: 'var(--accent-blue)',
                                    transition: 'width 0.4s ease-out'
                                }} 
                            />
                        </div>

                        {/* Node Contribution Breakdown */}
                        {Object.keys(nodeStats).length > 0 && (
                            <div style={{ marginTop: '12px', padding: '12px', background: 'var(--surface-sunken)', borderRadius: '8px' }}>
                                <div style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '8px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                                    Cluster Contribution
                                </div>
                                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                                    {Object.entries(nodeStats).map(([nodeId, count]) => (
                                        <div key={nodeId} className="badge alive" style={{ padding: '4px 8px' }}>
                                            <span className="badge-dot" />
                                            Node {nodeId.slice(0, 4)}: <strong>{count} pages</strong>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {completedDocs.length > 0 && (
                <div>
                    {completedDocs.map((doc) => {
                        const isExpanded = expandedJobId === doc.jobId;
                        const previewText = doc.text.length > 300 ? doc.text.slice(0, 300) + '…' : doc.text;

                        return (
                            <div key={doc.jobId} className="card result-card">
                                <div className="card-header" style={{ cursor: 'pointer' }} onClick={() => toggleExpand(doc.jobId)}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                                        <div
                                            className="list-item-icon"
                                            style={{
                                                background: 'var(--accent-green-dim)',
                                                color: 'var(--accent-green)',
                                                flexShrink: 0,
                                            }}
                                        >
                                            <IconCheck width={18} height={18} />
                                        </div>
                                        <div>
                                            <div className="card-title">Job {doc.jobId.slice(0, 8)}…</div>
                                            <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '2px' }}>
                                                {doc.totalPages} pages • Consensus verified
                                            </div>
                                        </div>
                                    </div>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <button
                                            className="btn btn-secondary btn-sm"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleCopy(doc.text, doc.jobId);
                                            }}
                                        >
                                            {copySuccess === doc.jobId ? '✓ Copied' : 'Copy Text'}
                                        </button>
                                        <span style={{ fontSize: '13px', color: 'var(--text-tertiary)' }}>
                                            {isExpanded ? '▲ Collapse' : '▼ Expand'}
                                        </span>
                                    </div>
                                </div>
                                <div className="card-body">
                                    {isExpanded ? (
                                        <div className="result-text-viewer">
                                            <pre className="result-text">{doc.text}</pre>
                                        </div>
                                    ) : (
                                        <div className="result-text-preview">
                                            <pre className="result-text" style={{ maxHeight: '120px', overflow: 'hidden' }}>{previewText}</pre>
                                        </div>
                                    )}
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default Results;
