import React, { useState } from 'react';
import { IconPower, IconServer, IconShield } from '../components/Icons';
import { swim } from '../../wailsjs/go/models';

interface SettingsProps {
    selfNode?: swim.Node;
    onKillNode: () => void;
}

const Settings: React.FC<SettingsProps> = ({ selfNode, onKillNode }) => {
    const [killed, setKilled] = useState(false);

    const handleKill = () => {
        if (window.confirm('Are you sure? This will stop the local SWIM heartbeat and simulate a node failure.')) {
            onKillNode();
            setKilled(true);
        }
    };

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Settings</h1>
                    <p className="page-subtitle">Node configuration and system controls</p>
                </div>
            </div>

            {/* Node Identity */}
            <div className="settings-section">
                <div className="settings-section-title">Node Identity</div>
                <div className="settings-row">
                    <span className="settings-row-label">Node ID</span>
                    <span className="settings-row-value">{selfNode?.id ?? '—'}</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">IP Address</span>
                    <span className="settings-row-value">{selfNode?.ip ?? '—'}</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">SWIM Port</span>
                    <span className="settings-row-value">{selfNode?.port ?? '—'}</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">RPC Port</span>
                    <span className="settings-row-value">{selfNode ? selfNode.port + 1 : '—'}</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">Status</span>
                    <span className="settings-row-value" style={{ color: killed ? 'var(--accent-red)' : 'var(--accent-green)' }}>
                        {killed ? 'Dead (Killed)' : selfNode?.status ?? '—'}
                    </span>
                </div>
            </div>

            {/* Network */}
            <div className="settings-section">
                <div className="settings-section-title">Network</div>
                <div className="card">
                    <div className="card-body" style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                        <div
                            className="list-item-icon"
                            style={{
                                background: 'rgba(6, 182, 212, 0.12)',
                                color: 'var(--accent-cyan)',
                                flexShrink: 0,
                            }}
                        >
                            <IconServer width={20} height={20} />
                        </div>
                        <div>
                            <div style={{ fontWeight: 600, fontSize: '14px', marginBottom: '4px' }}>SWIM Protocol</div>
                            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                                The SWIM gossip protocol is used for failure detection and membership management. Nodes periodically
                                ping random peers and propagate membership state changes through the cluster.
                            </div>
                        </div>
                    </div>
                </div>
                <div className="card" style={{ marginTop: '12px' }}>
                    <div className="card-body" style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                        <div
                            className="list-item-icon"
                            style={{
                                background: 'rgba(168, 85, 247, 0.12)',
                                color: 'var(--accent-purple)',
                                flexShrink: 0,
                            }}
                        >
                            <IconShield width={20} height={20} />
                        </div>
                        <div>
                            <div style={{ fontWeight: 600, fontSize: '14px', marginBottom: '4px' }}>Privacy-First Architecture</div>
                            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                                All processing occurs within your local network boundary. No data is sent to external servers.
                                Documents are processed using Tesseract OCR directly on worker nodes.
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Danger Zone */}
            <div className="settings-section">
                <div className="settings-section-title">Danger Zone</div>
                <div className="danger-zone">
                    <div className="danger-zone-title">Kill This Node</div>
                    <div className="danger-zone-desc">
                        Stops the SWIM heartbeat on this node, simulating a node failure. Other nodes will detect this
                        and mark you as Dead within a few seconds. Use for testing fault tolerance.
                    </div>
                    <button
                        className="btn btn-danger"
                        onClick={handleKill}
                        disabled={killed}
                    >
                        <IconPower className="btn-icon" />
                        {killed ? 'Node Killed' : 'Kill Node'}
                    </button>
                </div>
            </div>

            {/* About */}
            <div className="settings-section">
                <div className="settings-section-title">About</div>
                <div className="settings-row">
                    <span className="settings-row-label">Application</span>
                    <span className="settings-row-value">DOCR: A Local First OCR Application built on the Work Stealing Algorithm</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">Architecture</span>
                    <span className="settings-row-value">Master-Worker (LAN)</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">OCR Engine</span>
                    <span className="settings-row-value">Tesseract (CGO)</span>
                </div>
                <div className="settings-row">
                    <span className="settings-row-label">Framework</span>
                    <span className="settings-row-value">Wails v2 + React</span>
                </div>
            </div>
        </div>
    );
};

export default Settings;
