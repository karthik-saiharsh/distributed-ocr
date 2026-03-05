import React from 'react';
import { IconResults, IconFileText, IconShield } from '../components/Icons';

const Results: React.FC = () => {
    // Placeholder — will be wired when backend exposes verified results API.
    // For now, shows an informative empty state.
    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Results</h1>
                    <p className="page-subtitle">View verified OCR outputs from processed documents</p>
                </div>
            </div>

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
        </div>
    );
};

export default Results;
