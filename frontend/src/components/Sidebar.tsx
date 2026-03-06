import React from 'react';
import {
    IconDashboard,
    IconCluster,
    IconJobs,
    IconResults,
    IconSettings,
} from './Icons';

export type Page = 'dashboard' | 'cluster' | 'jobs' | 'results' | 'settings';

interface SidebarProps {
    activePage: Page;
    onNavigate: (page: Page) => void;
    selfNodeIP?: string;
    selfNodePort?: number;
}

const navItems: { id: Page; label: string; icon: React.ReactNode }[] = [
    { id: 'dashboard', label: 'Dashboard', icon: <IconDashboard className="nav-icon" /> },
    { id: 'cluster', label: 'Cluster', icon: <IconCluster className="nav-icon" /> },
    { id: 'jobs', label: 'Jobs', icon: <IconJobs className="nav-icon" /> },
    { id: 'results', label: 'Results', icon: <IconResults className="nav-icon" /> },
    { id: 'settings', label: 'Settings', icon: <IconSettings className="nav-icon" /> },
];

const Sidebar: React.FC<SidebarProps> = ({ activePage, onNavigate, selfNodeIP, selfNodePort }) => {
    return (
        <aside className="sidebar">
            {/* Logo */}
            <div className="sidebar-logo">
                <div className="logo-icon">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                    </svg>
                </div>
                <div>
                    <div className="logo-text">DOCR</div>
                    <div className="logo-sub">A Local First OCR Application</div>
                </div>
            </div>

            {/* Navigation */}
            <nav className="sidebar-nav">
                {navItems.map((item) => (
                    <div
                        key={item.id}
                        className={`sidebar-nav-item${activePage === item.id ? ' active' : ''}`}
                        onClick={() => onNavigate(item.id)}
                    >
                        {item.icon}
                        {item.label}
                    </div>
                ))}
            </nav>

            {/* Footer — self node */}
            <div className="sidebar-footer">
                <div className="self-node-badge">
                    <span className="self-node-dot" />
                    <div className="self-node-info">
                        <div className="self-node-label">This Node</div>
                        <div className="self-node-addr">
                            {selfNodeIP ? `${selfNodeIP}:${selfNodePort}` : 'Initializing…'}
                        </div>
                    </div>
                </div>
            </div>
        </aside>
    );
};

export default Sidebar;
