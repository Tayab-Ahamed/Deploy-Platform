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
        // Resume polling for any in-progress jobs
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
    <div style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      {/* Top nav */}
      <nav style={{
        borderBottom: '1px solid var(--border)',
        padding: '0 24px',
        height: 52,
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        background: 'var(--bg2)',
        position: 'sticky',
        top: 0,
        zIndex: 10,
      }}>
        <span style={{ fontSize: 18 }}>🚀</span>
        <span style={{ fontWeight: 700, fontSize: 15, letterSpacing: '-0.01em' }}>Deploy Platform</span>
        <span style={{
          fontSize: 10, fontWeight: 600, letterSpacing: '0.08em',
          background: 'rgba(124,111,247,0.15)', color: '#7c6ff7',
          padding: '2px 7px', borderRadius: 4, textTransform: 'uppercase',
        }}>LOCAL</span>
        <div style={{ flex: 1 }} />
        <div style={{ fontSize: 12, color: 'var(--muted)', display: 'flex', gap: 16 }}>
          <span>Total: <b style={{ color: 'var(--text)' }}>{stats.total}</b></span>
          <span>Live: <b style={{ color: '#22c77a' }}>{stats.live}</b></span>
          {stats.building > 0 && <span>Building: <b style={{ color: '#f0a500' }}>{stats.building}</b></span>}
          {stats.failed > 0 && <span>Failed: <b style={{ color: '#e05454' }}>{stats.failed}</b></span>}
        </div>
      </nav>

      {/* Main content */}
      <main style={{ maxWidth: 820, margin: '0 auto', padding: '32px 20px' }}>

        {/* Hero */}
        <div style={{ marginBottom: 32 }}>
          <h1 style={{ fontSize: 28, fontWeight: 700, letterSpacing: '-0.02em', marginBottom: 6 }}>
            Deployments
          </h1>
          <p style={{ color: 'var(--muted)', fontSize: 14 }}>
            Zip your project, give it a name, and deploy it as a Docker container on Kubernetes.
          </p>
        </div>

        {/* Deploy form */}
        <DeployForm onDeployStarted={handleDeployStarted} />

        {/* Deployment list */}
        {loading ? (
          <div style={{ textAlign: 'center', color: 'var(--muted)', padding: '48px 0' }}>
            Loading deployments...
          </div>
        ) : deploys.length === 0 ? (
          <div style={{
            textAlign: 'center', color: 'var(--muted)', padding: '48px 0',
            border: '1px dashed var(--border)', borderRadius: 'var(--radius-lg)',
          }}>
            <div style={{ fontSize: 32, marginBottom: 12 }}>📭</div>
            <div style={{ fontWeight: 500, marginBottom: 6 }}>No deployments yet</div>
            <div style={{ fontSize: 12 }}>Upload a zip above to get started</div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {deploys.map(job => (
              <DeployCard key={job.id} job={job} />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
