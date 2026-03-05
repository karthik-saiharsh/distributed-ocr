import React from 'react';
import StatusBadge from './StatusBadge';
import { swim } from '../../wailsjs/go/models';

interface NodeCardProps {
    node: swim.Node;
    isSelf?: boolean;
}

function timeAgo(lastUpdated: any): string {
    if (!lastUpdated) return 'unknown';
    const date = new Date(lastUpdated);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffS = Math.floor(diffMs / 1000);
    if (diffS < 5) return 'just now';
    if (diffS < 60) return `${diffS}s ago`;
    const diffM = Math.floor(diffS / 60);
    if (diffM < 60) return `${diffM}m ago`;
    const diffH = Math.floor(diffM / 60);
    return `${diffH}h ago`;
}

function truncateId(id: string, len = 12): string {
    if (id.length <= len) return id;
    return id.substring(0, len) + '…';
}

const NodeCard: React.FC<NodeCardProps> = ({ node, isSelf = false }) => {
    return (
        <div className={`node-card animate-fade-in${isSelf ? ' self' : ''}`}>
            <div className="node-card-header">
                <span className="node-card-id" title={node.id}>
                    {truncateId(node.id)}
                </span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    {isSelf && <span className="node-self-tag">YOU</span>}
                    <StatusBadge status={node.status} />
                </div>
            </div>
            <div className="node-card-body">
                <div className="node-card-row">
                    <span className="node-card-label">Address</span>
                    <span className="node-card-value">{node.ip}:{node.port}</span>
                </div>
                <div className="node-card-row">
                    <span className="node-card-label">Last Seen</span>
                    <span className="node-card-value">{timeAgo(node.lastUpdated)}</span>
                </div>
            </div>
        </div>
    );
};

export default NodeCard;
