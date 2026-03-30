import React, { useState } from 'react'
import StatusBadge from './StatusBadge.jsx'
import LogViewer from './LogViewer.jsx'

export default function DeployCard({ job }) {
  const [expanded, setExpanded] = useState(job.status === 'building' || job.status === 'deploying')

  const ts = new Date(job.created_at).toLocaleString()

  return (
    <div style={{
      background: 'var(--bg2)',
      border: '1px solid var(--border)',
      borderRadius: 'var(--radius-lg)',
      padding: '16px 20px',
      transition: 'border-color 0.2s',
    }}
    onMouseEnter={e => e.currentTarget.style.borderColor = 'var(--border2)'}
    onMouseLeave={e => e.currentTarget.style.borderColor = 'var(--border)'}
    >
      {/* Header row */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1, minWidth: 0 }}>
          <span style={{
            width: 32, height: 32, borderRadius: 8,
            background: 'linear-gradient(135deg, #7c6ff7 0%, #4fa8e0 100%)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 14, flexShrink: 0,
          }}>
            {job.app_name?.[0]?.toUpperCase() ?? '?'}
          </span>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontWeight: 600, fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {job.app_name}
            </div>
            <div style={{ color: 'var(--muted)', fontSize: 11 }}>
              {job.id?.slice(0, 8)} · {ts}
            </div>
          </div>
        </div>

        <StatusBadge status={job.status} />

        {job.url && (
          <a href={job.url} target="_blank" rel="noreferrer" style={{
            fontSize: 12, color: 'var(--green)', display: 'flex', alignItems: 'center', gap: 4,
          }}>
            {job.url} ↗
          </a>
        )}

        <button
          onClick={() => setExpanded(v => !v)}
          style={{
            background: 'transparent', border: '1px solid var(--border)',
            borderRadius: 6, padding: '3px 10px', color: 'var(--muted)',
            fontSize: 12, cursor: 'pointer',
          }}
        >
          {expanded ? 'Hide logs' : 'Show logs'}
        </button>
      </div>

      {expanded && <LogViewer logs={job.logs} />}
    </div>
  )
}
