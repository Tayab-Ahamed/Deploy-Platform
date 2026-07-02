import React from 'react'

const CONFIG = {
  queued:    { label: 'Queued',     color: 'var(--muted)', bg: 'rgba(130,130,146,0.08)', dot: 'var(--muted)' },
  building:  { label: 'Building',   color: 'var(--amber)', bg: 'var(--amber-glow)',      dot: 'var(--amber)', pulse: true },
  deploying: { label: 'Deploying',  color: 'var(--blue)',  bg: 'var(--blue-glow)',       dot: 'var(--blue)',  pulse: true },
  live:      { label: 'Live',       color: 'var(--green)', bg: 'var(--green-glow)',      dot: 'var(--green)' },
  failed:    { label: 'Failed',     color: 'var(--red)',   bg: 'var(--red-glow)',        dot: 'var(--red)' },
}

export default function StatusBadge({ status }) {
  const c = CONFIG[status] || CONFIG.queued
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 5,
      padding: '3px 10px', borderRadius: 20,
      background: c.bg, color: c.color,
      fontSize: 12, fontWeight: 500, letterSpacing: '0.02em',
    }}>
      <span style={{
        width: 6, height: 6, borderRadius: '50%',
        background: c.dot,
        animation: c.pulse ? 'pulse 1.4s ease-in-out infinite' : 'none',
      }} />
      {c.label}
      <style>{`
        @keyframes pulse {
          0%,100% { opacity: 1; transform: scale(1); }
          50% { opacity: 0.4; transform: scale(0.8); }
        }
      `}</style>
    </span>
  )
}
