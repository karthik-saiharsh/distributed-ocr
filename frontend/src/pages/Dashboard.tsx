import React from 'react';
import StatCard from '../components/StatCard';
import NodeCard from '../components/NodeCard';
import {
    IconCluster,
    IconCheck,
    IconQueue,
    IconActivity,
    IconScan,
    IconUpload,
    IconServer,
} from '../components/Icons';
import { swim } from '../../wailsjs/go/models';

interface DashboardProps {
    nodes: swim.Node[];
    selfNodeId?: string;
    scanning: boolean;
    onScan: () => void;
    onUpload: () => void;
    onNavigate: (page: string) => void;
}

const Dashboard: React.FC<DashboardProps> = ({
    nodes,
    selfNodeId,
    scanning,
    onScan,
    onUpload,
    onNavigate,
}) => {
    const aliveCount = nodes.filter((n) => n.status === 'Alive').length;
    const suspectCount = nodes.filter((n) => n.status === 'Suspect').length;
    const deadCount = nodes.filter((n) => n.status === 'Dead').length;
    const totalNodes = nodes.length;

    // Health ring SVG
    const radius = 64;
    const circumference = 2 * Math.PI * radius;
    const total = Math.max(totalNodes, 1);
    const aliveArc = (aliveCount / total) * circumference;
    const suspectArc = (suspectCount / total) * circumference;
    const deadArc = (deadCount / total) * circumference;
    const aliveOffset = 0;
    const suspectOffset = aliveArc;
    const deadOffset = aliveArc + suspectArc;

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Dashboard</h1>
                    <p className="page-subtitle">Cluster overview and quick actions</p>
                </div>
                <div className="page-actions">
                    <button className="btn btn-secondary" onClick={onScan} disabled={scanning}>
                        <IconScan className={`btn-icon${scanning ? ' animate-spin' : ''}`} />
                        {scanning ? 'Scanning…' : 'Scan Network'}
                    </button>
                    <button className="btn btn-primary" onClick={onUpload}>
                        <IconUpload className="btn-icon" />
                        Upload Document
                    </button>
                </div>
            </div>

            {/* Stat Cards */}
            <div className="stat-grid">
                <StatCard
                    icon={<IconCluster width={20} height={20} />}
                    value={totalNodes}
                    label="Total Nodes"
                    color="blue"
                />
                <StatCard
                    icon={<IconServer width={20} height={20} />}
                    value={aliveCount}
                    label="Alive Nodes"
                    color="green"
                />
                <StatCard
                    icon={<IconCheck width={20} height={20} />}
                    value={0}
                    label="Jobs Completed"
                    color="purple"
                />
                <StatCard
                    icon={<IconQueue width={20} height={20} />}
                    value={0}
                    label="Queue Depth"
                    color="amber"
                />
            </div>

            {/* Body — two columns */}
            <div className="grid-2 section-gap">
                {/* Cluster Health */}
                <div className="card">
                    <div className="card-header">
                        <span className="card-title">Cluster Health</span>
                        <button className="btn btn-ghost btn-sm" onClick={() => onNavigate('cluster')}>
                            View All
                        </button>
                    </div>
                    <div className="card-body">
                        {totalNodes === 0 ? (
                            <div className="empty-state" style={{ padding: '24px' }}>
                                <div className="empty-state-icon">
                                    <IconCluster width={28} height={28} />
                                </div>
                                <div className="empty-state-title">No nodes yet</div>
                                <div className="empty-state-desc">
                                    Press "Scan Network" to discover nodes on your LAN.
                                </div>
                            </div>
                        ) : (
                            <div className="health-ring-container">
                                <div className="health-ring">
                                    <svg width="160" height="160" viewBox="0 0 160 160">
                                        {/* Background ring */}
                                        <circle
                                            cx="80"
                                            cy="80"
                                            r={radius}
                                            fill="none"
                                            stroke="rgba(255,255,255,0.04)"
                                            strokeWidth="14"
                                        />
                                        {/* Alive arc */}
                                        {aliveCount > 0 && (
                                            <circle
                                                cx="80"
                                                cy="80"
                                                r={radius}
                                                fill="none"
                                                stroke="var(--accent-green)"
                                                strokeWidth="14"
                                                strokeDasharray={`${aliveArc} ${circumference - aliveArc}`}
                                                strokeDashoffset={-aliveOffset}
                                                strokeLinecap="round"
                                            />
                                        )}
                                        {/* Suspect arc */}
                                        {suspectCount > 0 && (
                                            <circle
                                                cx="80"
                                                cy="80"
                                                r={radius}
                                                fill="none"
                                                stroke="var(--accent-amber)"
                                                strokeWidth="14"
                                                strokeDasharray={`${suspectArc} ${circumference - suspectArc}`}
                                                strokeDashoffset={-suspectOffset}
                                                strokeLinecap="round"
                                            />
                                        )}
                                        {/* Dead arc */}
                                        {deadCount > 0 && (
                                            <circle
                                                cx="80"
                                                cy="80"
                                                r={radius}
                                                fill="none"
                                                stroke="var(--accent-red)"
                                                strokeWidth="14"
                                                strokeDasharray={`${deadArc} ${circumference - deadArc}`}
                                                strokeDashoffset={-deadOffset}
                                                strokeLinecap="round"
                                            />
                                        )}
                                    </svg>
                                    <div className="health-ring-label">
                                        <div className="health-ring-value">{aliveCount}</div>
                                        <div className="health-ring-text">alive</div>
                                    </div>
                                </div>
                                <div className="health-legend">
                                    <div className="health-legend-item">
                                        <span className="health-legend-dot" style={{ background: 'var(--accent-green)' }} />
                                        <span>Alive</span>
                                        <span className="health-legend-count">{aliveCount}</span>
                                    </div>
                                    <div className="health-legend-item">
                                        <span className="health-legend-dot" style={{ background: 'var(--accent-amber)' }} />
                                        <span>Suspect</span>
                                        <span className="health-legend-count">{suspectCount}</span>
                                    </div>
                                    <div className="health-legend-item">
                                        <span className="health-legend-dot" style={{ background: 'var(--accent-red)' }} />
                                        <span>Dead</span>
                                        <span className="health-legend-count">{deadCount}</span>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Recent Activity */}
                <div className="card">
                    <div className="card-header">
                        <span className="card-title">Recent Activity</span>
                        <IconActivity width={16} height={16} style={{ color: 'var(--text-tertiary)' }} />
                    </div>
                    <div className="card-body">
                        <div className="activity-item">
                            <span className="activity-dot" style={{ background: 'var(--accent-green)' }} />
                            <div style={{ flex: 1 }}>
                                <span className="activity-text">
                                    <strong>System started</strong> — SWIM discovery protocol initialized
                                </span>
                                <div className="activity-time">Just now</div>
                            </div>
                        </div>
                        <div className="activity-item">
                            <span className="activity-dot" style={{ background: 'var(--accent-blue)' }} />
                            <div style={{ flex: 1 }}>
                                <span className="activity-text">
                                    <strong>Self node registered</strong> — ready to join cluster
                                </span>
                                <div className="activity-time">Just now</div>
                            </div>
                        </div>
                        {nodes.length > 1 && (
                            <div className="activity-item">
                                <span className="activity-dot" style={{ background: 'var(--accent-purple)' }} />
                                <div style={{ flex: 1 }}>
                                    <span className="activity-text">
                                        <strong>{nodes.length - 1} peer(s)</strong> discovered via network scan
                                    </span>
                                    <div className="activity-time">Recently</div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Quick node preview */}
            {nodes.length > 0 && (
                <div className="section-gap">
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px' }}>
                        <h3 style={{ fontSize: '16px', fontWeight: 600 }}>Active Nodes</h3>
                        <button className="btn btn-ghost btn-sm" onClick={() => onNavigate('cluster')}>
                            View all →
                        </button>
                    </div>
                    <div className="node-grid">
                        {nodes.slice(0, 3).map((node) => (
                            <NodeCard
                                key={node.id}
                                node={node}
                                isSelf={node.id === selfNodeId}
                            />
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default Dashboard;
