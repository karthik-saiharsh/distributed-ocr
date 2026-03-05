import React, { useState } from 'react';
import NodeCard from '../components/NodeCard';
import { IconScan, IconCluster } from '../components/Icons';
import { swim } from '../../wailsjs/go/models';

interface ClusterProps {
    nodes: swim.Node[];
    selfNodeId?: string;
    scanning: boolean;
    onScan: () => void;
}

type StatusFilter = 'all' | 'Alive' | 'Suspect' | 'Dead';

const Cluster: React.FC<ClusterProps> = ({ nodes, selfNodeId, scanning, onScan }) => {
    const [filter, setFilter] = useState<StatusFilter>('all');

    const filteredNodes =
        filter === 'all' ? nodes : nodes.filter((n) => n.status === filter);

    const aliveCount = nodes.filter((n) => n.status === 'Alive').length;
    const suspectCount = nodes.filter((n) => n.status === 'Suspect').length;
    const deadCount = nodes.filter((n) => n.status === 'Dead').length;

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Cluster</h1>
                    <p className="page-subtitle">Manage and monitor network nodes</p>
                </div>
                <div className="page-actions">
                    <button className="btn btn-primary" onClick={onScan} disabled={scanning}>
                        <IconScan className={`btn-icon${scanning ? ' animate-spin' : ''}`} />
                        {scanning ? 'Scanning…' : 'Scan for Nodes'}
                    </button>
                </div>
            </div>

            {/* Toolbar */}
            <div className="toolbar">
                <div className="filter-pills">
                    <button
                        className={`filter-pill${filter === 'all' ? ' active' : ''}`}
                        onClick={() => setFilter('all')}
                    >
                        All <span className="count-badge" style={{ marginLeft: 6 }}>{nodes.length}</span>
                    </button>
                    <button
                        className={`filter-pill${filter === 'Alive' ? ' active' : ''}`}
                        onClick={() => setFilter('Alive')}
                    >
                        Alive <span className="count-badge" style={{ marginLeft: 6 }}>{aliveCount}</span>
                    </button>
                    <button
                        className={`filter-pill${filter === 'Suspect' ? ' active' : ''}`}
                        onClick={() => setFilter('Suspect')}
                    >
                        Suspect <span className="count-badge" style={{ marginLeft: 6 }}>{suspectCount}</span>
                    </button>
                    <button
                        className={`filter-pill${filter === 'Dead' ? ' active' : ''}`}
                        onClick={() => setFilter('Dead')}
                    >
                        Dead <span className="count-badge" style={{ marginLeft: 6 }}>{deadCount}</span>
                    </button>
                </div>
            </div>

            {/* Node Grid */}
            {filteredNodes.length === 0 ? (
                <div className="card" style={{ marginTop: '8px' }}>
                    <div className="empty-state">
                        <div className="empty-state-icon">
                            <IconCluster width={28} height={28} />
                        </div>
                        <div className="empty-state-title">
                            {filter === 'all' ? 'No nodes discovered' : `No ${filter.toLowerCase()} nodes`}
                        </div>
                        <div className="empty-state-desc">
                            {filter === 'all'
                                ? 'Press "Scan for Nodes" to discover peers on your local network.'
                                : 'Try changing the filter to see other nodes.'}
                        </div>
                        {filter === 'all' && (
                            <button className="btn btn-primary" onClick={onScan} disabled={scanning}>
                                <IconScan className={`btn-icon${scanning ? ' animate-spin' : ''}`} />
                                {scanning ? 'Scanning…' : 'Scan Network'}
                            </button>
                        )}
                    </div>
                </div>
            ) : (
                <div className="node-grid">
                    {filteredNodes.map((node) => (
                        <NodeCard
                            key={node.id}
                            node={node}
                            isSelf={node.id === selfNodeId}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

export default Cluster;
