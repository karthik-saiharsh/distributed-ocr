import React from 'react';

interface StatCardProps {
    icon: React.ReactNode;
    value: string | number;
    label: string;
    color: 'blue' | 'green' | 'amber' | 'red' | 'purple' | 'cyan';
}

const StatCard: React.FC<StatCardProps> = ({ icon, value, label, color }) => {
    return (
        <div className="stat-card animate-fade-in">
            <div className={`stat-card-icon ${color}`}>
                {icon}
            </div>
            <div className="stat-card-value">{value}</div>
            <div className="stat-card-label">{label}</div>
        </div>
    );
};

export default StatCard;
