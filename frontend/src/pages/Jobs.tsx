import React from 'react';
import { IconUpload, IconFileText, IconQueue, IconCheck } from '../components/Icons';
import type { JobInfo } from '../App';

interface JobsProps {
    jobs: JobInfo[];
    onUpload: () => void;
}

const Jobs: React.FC<JobsProps> = ({ jobs, onUpload }) => {
    const getProgress = (job: JobInfo): number => {
        if (job.status === 'completed') return 100;
        if (job.percentage !== undefined) return Math.round(job.percentage);
        if (job.totalPages === 0) return 0;
        return Math.round((job.completedPages / job.totalPages) * 100);
    };

    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Jobs</h1>
                    <p className="page-subtitle">Upload documents and track processing</p>
                </div>
            </div>

            {/* Upload Zone */}
            <div className="upload-zone" onClick={onUpload}>
                <div className="upload-zone-icon">
                    <IconUpload width={28} height={28} />
                </div>
                <div className="upload-zone-title">Upload Document</div>
                <div className="upload-zone-subtitle">
                    Click to upload a PDF for distributed OCR processing
                </div>
                <button className="btn btn-primary" style={{ marginTop: '8px' }} onClick={(e) => { e.stopPropagation(); onUpload(); }}>
                    <IconUpload className="btn-icon" />
                    Choose File
                </button>
            </div>

            {/* Job List */}
            <div className="section-gap">
                <h3 style={{ fontSize: '16px', fontWeight: 600, marginBottom: '16px' }}>
                    Job Queue {jobs.length > 0 && <span className="count-badge" style={{ marginLeft: 8 }}>{jobs.length}</span>}
                </h3>

                {jobs.length === 0 ? (
                    <div className="card">
                        <div className="empty-state">
                            <div className="empty-state-icon">
                                <IconQueue width={28} height={28} />
                            </div>
                            <div className="empty-state-title">No jobs yet</div>
                            <div className="empty-state-desc">
                                Upload a document to start processing it across the cluster.
                            </div>
                        </div>
                    </div>
                ) : (
                    jobs.map((job) => {
                        const progress = getProgress(job);
                        return (
                            <div key={job.id} className="list-item">
                                <div
                                    className="list-item-icon"
                                    style={{
                                        background:
                                            job.status === 'completed'
                                                ? 'var(--accent-green-dim)'
                                                : job.status === 'processing'
                                                    ? 'rgba(99, 102, 241, 0.12)'
                                                    : 'var(--accent-amber-dim)',
                                        color:
                                            job.status === 'completed'
                                                ? 'var(--accent-green)'
                                                : job.status === 'processing'
                                                    ? 'var(--accent-blue)'
                                                    : 'var(--accent-amber)',
                                    }}
                                >
                                    {job.status === 'completed' ? (
                                        <IconCheck width={20} height={20} />
                                    ) : (
                                        <IconFileText width={20} height={20} />
                                    )}
                                </div>
                                <div className="list-item-content">
                                    <div className="list-item-title">{job.name}</div>
                                    <div className="list-item-sub">
                                        {job.totalPages} pages • {job.completedPages}/{job.totalPages} verified •{' '}
                                        {job.status === 'completed' ? 'Completed' : job.status === 'processing' ? 'Processing' : 'Queued'}
                                    </div>
                                    {job.status !== 'completed' && (
                                        <div className="progress-bar-track" style={{ marginTop: 8 }}>
                                            <div
                                                className="progress-bar-fill"
                                                style={{ width: `${progress}%`, transition: 'width 0.3s ease-in-out' }}
                                            />
                                        </div>
                                    )}
                                    {job.status === 'completed' && (
                                        <div className="progress-bar-track" style={{ marginTop: 8 }}>
                                            <div
                                                className="progress-bar-fill"
                                                style={{ width: '100%', background: 'var(--accent-green)' }}
                                            />
                                        </div>
                                    )}
                                </div>
                                <div className="list-item-meta">
                                    <span className={`badge ${job.status === 'completed' ? 'alive' : job.status === 'processing' ? 'suspect' : 'dead'}`}>
                                        <span className="badge-dot" />
                                        {job.status === 'completed' ? 'Done' : job.status === 'processing' ? `${progress}%` : 'Queued'}
                                    </span>
                                </div>
                            </div>
                        );
                    })
                )}
            </div>
        </div>
    );
};

export default Jobs;
