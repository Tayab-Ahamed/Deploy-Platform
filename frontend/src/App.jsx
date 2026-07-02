import React, { useState, useEffect, useCallback, useRef } from 'react'
import DeployForm from './components/DeployForm.jsx'
import DeployCard from './components/DeployCard.jsx'
import { listDeploys, getStatus } from './api/client.js'

export default function App() {
  const [deploys, setDeploys] = useState([])
  const [loading, setLoading] = useState(true)
  const pollRefs = useRef({})

  // Fetch all deployments on mount
  useEffect(() => {
    listDeploys()
      .then(data => {
        setDeploys(Array.isArray(data) ? data : [])
        ;(Array.isArray(data) ? data : []).forEach(job => {
          if (['queued', 'building', 'deploying'].includes(job.status)) {
            startPolling(job.id)
          }
        })
      })
      .catch(() => setDeploys([]))
      .finally(() => setLoading(false))
  }, [])

  const startPolling = useCallback((id) => {
    if (pollRefs.current[id]) return // already polling
    const interval = setInterval(async () => {
      try {
        const job = await getStatus(id)
        setDeploys(prev => {
          const exists = prev.find(d => d.id === id)
          if (exists) return prev.map(d => d.id === id ? job : d)
          return [job, ...prev]
        })
        if (job.status === 'live' || job.status === 'failed') {
          clearInterval(interval)
          delete pollRefs.current[id]
        }
      } catch {
        clearInterval(interval)
        delete pollRefs.current[id]
      }
    }, 2000)
    pollRefs.current[id] = interval
  }, [])

  const handleDeployStarted = useCallback((id) => {
    // Add placeholder immediately
    setDeploys(prev => [{
      id,
      app_name: '...',
      status: 'queued',
      logs: ['Job queued, waiting for build worker...'],
      url: '',
      created_at: new Date().toISOString(),
    }, ...prev])
    startPolling(id)
  }, [startPolling])

  const handleDeployDeleted = useCallback((id) => {
    // Clear polling if running
    if (pollRefs.current[id]) {
      clearInterval(pollRefs.current[id])
      delete pollRefs.current[id]
    }
    setDeploys(prev => prev.filter(d => d.id !== id))
  }, [])

  // Clean up on unmount
  useEffect(() => {
    return () => Object.values(pollRefs.current).forEach(clearInterval)
  }, [])

  const stats = {
    total: deploys.length,
    live: deploys.filter(d => d.status === 'live').length,
    building: deploys.filter(d => ['queued','building','deploying'].includes(d.status)).length,
    failed: deploys.filter(d => d.status === 'failed').length,
  }

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', color: 'var(--text)' }}>
      {/* Top nav */}
      <nav 
        className="glass-panel animate-fade-in" 
        style={{
          borderBottom: '1px solid var(--border)',
          borderLeft: 'none', borderRight: 'none', borderTop: 'none',
          padding: '0 32px',
          height: 60,
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          position: 'sticky',
          top: 0,
          zIndex: 10,
        }}
      >
        <span style={{ fontSize: 20, filter: 'drop-shadow(0 0 6px rgba(99, 102, 241, 0.4))' }}>🚀</span>
        <span style={{ fontWeight: 800, fontSize: 16, letterSpacing: '-0.02em', background: 'linear-gradient(to right, #fff, #a5b4fc)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
          Deploy Platform
        </span>
        <span style={{
          fontSize: 9, fontWeight: 700, letterSpacing: '0.08em',
          background: 'var(--accent-glow)', color: '#818cf8',
          padding: '3px 8px', borderRadius: 4, textTransform: 'uppercase',
          border: '1px solid rgba(99, 102, 241, 0.2)',
        }}>
          WSL2 Cluster
        </span>
        <div style={{ flex: 1 }} />
        <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--muted)', display: 'flex', gap: 20 }}>
          <span>Total: <b style={{ color: 'var(--text)' }}>{stats.total}</b></span>
          <span>Live: <b style={{ color: 'var(--green)' }}>{stats.live}</b></span>
          {stats.building > 0 && <span style={{ animation: 'pulse 1.5s infinite' }}>Building: <b style={{ color: 'var(--amber)' }}>{stats.building}</b></span>}
          {stats.failed > 0 && <span>Failed: <b style={{ color: 'var(--red)' }}>{stats.failed}</b></span>}
        </div>
      </nav>

      {/* Main content */}
      <main className="animate-fade-in" style={{ maxWidth: 880, margin: '0 auto', padding: '40px 24px' }}>

        {/* Hero */}
        <div style={{ marginBottom: 36, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', flexWrap: 'wrap', gap: 12 }}>
          <div>
            <h1 style={{ fontSize: 32, fontWeight: 800, letterSpacing: '-0.03em', marginBottom: 6 }}>
              Applications
            </h1>
            <p style={{ color: 'var(--muted)', fontSize: 14 }}>
              Push package folders to build containers and deploy instantly to Kubernetes.
            </p>
          </div>
        </div>

        {/* Deploy form */}
        <DeployForm onDeployStarted={handleDeployStarted} />

        {/* Deployment list */}
        {loading ? (
          <div style={{ textAlign: 'center', color: 'var(--muted)', padding: '64px 0', fontSize: 13 }}>
            ⏳ Loading cluster state...
          </div>
        ) : deploys.length === 0 ? (
          <div className="glass-panel" style={{
            textAlign: 'center', color: 'var(--muted)', padding: '64px 32px',
            borderRadius: 'var(--radius-lg)', border: '1px dashed var(--border)',
          }}>
            <div style={{ fontSize: 40, marginBottom: 16 }}>📭</div>
            <div style={{ fontWeight: 700, color: 'var(--text)', fontSize: 16, marginBottom: 6 }}>No active deployments</div>
            <div style={{ fontSize: 13, color: 'var(--muted)' }}>Upload a zip file above to roll out your first application.</div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {deploys.map(job => (
              <DeployCard key={job.id} job={job} onDeployDeleted={handleDeployDeleted} />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
