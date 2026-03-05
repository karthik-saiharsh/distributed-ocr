import React, { useState } from 'react';
import { IconUpload, IconFileText, IconQueue, IconCheck } from '../components/Icons';

interface Job {
    id: string;
    name: string;
    pages: number;
    status: 'processing' | 'queued' | 'completed';
    progress: number; // 0-100
}

interface JobsProps {
    onUpload: () => void;
}

const Jobs: React.FC<JobsProps> = ({ onUpload }) => {
    // Track jobs locally. In a full integration, this would come from backend events.
    const [jobs, setJobs] = useState<Job[]>([]);
    const [uploadCount, setUploadCount] = useState(0);

    const handleUpload = () => {
        onUpload();
        const newJob: Job = {
            id: `job-${Date.now()}`,
            name: `Document ${uploadCount + 1}`,
            pages: 5,
            status: 'queued',
            progress: 0,
        };
        setJobs((prev) => [newJob, ...prev]);
        setUploadCount((c) => c + 1);
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
            <div className="upload-zone" onClick={handleUpload}>
                <div className="upload-zone-icon">
                    <IconUpload width={28} height={28} />
                </div>
                <div className="upload-zone-title">Upload Document</div>
                <div className="upload-zone-subtitle">
                    Click to upload a PDF for distributed OCR processing
                </div>
                <button className="btn btn-primary" style={{ marginTop: '8px' }} onClick={(e) => { e.stopPropagation(); handleUpload(); }}>
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
                    jobs.map((job) => (
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
                                    {job.pages} pages • {job.status === 'completed' ? 'Completed' : job.status === 'processing' ? 'Processing' : 'Queued'}
                                </div>
                                {job.status !== 'completed' && (
                                    <div className="progress-bar-track" style={{ marginTop: 8 }}>
                                        <div
                                            className="progress-bar-fill"
                                            style={{ width: `${job.progress}%` }}
                                        />
                                    </div>
                                )}
                            </div>
                            <div className="list-item-meta">
                                <span className={`badge ${job.status === 'completed' ? 'alive' : job.status === 'processing' ? 'alive' : 'suspect'}`}>
                                    <span className="badge-dot" />
                                    {job.status === 'completed' ? 'Done' : job.status === 'processing' ? 'Active' : 'Queued'}
                                </span>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default Jobs;
