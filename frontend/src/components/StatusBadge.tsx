import React from 'react';

interface StatusBadgeProps {
    status: string;
}

const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
    const normalized = status.toLowerCase();
    const variant =
        normalized === 'alive' ? 'alive' :
            normalized === 'suspect' ? 'suspect' : 'dead';

    return (
        <span className={`badge ${variant}`}>
            <span className="badge-dot" />
            {status}
        </span>
    );
};

export default StatusBadge;
