import React from 'react'

const CONFIG = {
  queued:    { label: 'Queued',     color: '#7c7c8a', bg: 'rgba(124,124,138,0.12)', dot: '#7c7c8a' },
  building:  { label: 'Building',   color: '#f0a500', bg: 'rgba(240,165,0,0.12)',   dot: '#f0a500', pulse: true },
  deploying: { label: 'Deploying',  color: '#4fa8e0', bg: 'rgba(79,168,224,0.12)',  dot: '#4fa8e0', pulse: true },
  live:      { label: 'Live',       color: '#22c77a', bg: 'rgba(34,199,122,0.12)',  dot: '#22c77a' },
  failed:    { label: 'Failed',     color: '#e05454', bg: 'rgba(224,84,84,0.12)',   dot: '#e05454' },
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
