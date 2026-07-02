import React, { useState, useEffect } from 'react'
import StatusBadge from './StatusBadge.jsx'
import LogViewer from './LogViewer.jsx'
import { deleteDeploy, scaleDeploy, getPods, getPodLogs } from '../api/client.js'

export default function DeployCard({ job, onDeployDeleted }) {
  const [expanded, setExpanded] = useState(job.status === 'building' || job.status === 'deploying')
  const [activeTab, setActiveTab] = useState('build') // 'build', 'pods', 'runtime', 'settings'
  const [pods, setPods] = useState([])
  const [podsLoading, setPodsLoading] = useState(false)
  const [selectedPod, setSelectedPod] = useState('')
  const [podLogs, setPodLogs] = useState('')
  const [podLogsLoading, setPodLogsLoading] = useState(false)
  const [replicas, setReplicas] = useState(job.replicas || 1)
  const [scaling, setScaling] = useState(false)

  const ts = new Date(job.created_at).toLocaleString()

  // Auto-set replicas state when job updates
  useEffect(() => {
    if (job.replicas) {
      setReplicas(job.replicas)
    }
  }, [job.replicas])

  // Fetch pods when tab switches to 'pods' or 'runtime'
  useEffect(() => {
    if (expanded && (activeTab === 'pods' || activeTab === 'runtime')) {
      loadPods()
    }
  }, [expanded, activeTab])

  // Fetch pod logs when selected pod changes
  useEffect(() => {
    if (expanded && activeTab === 'runtime' && selectedPod) {
      loadPodLogs(selectedPod)
    }
  }, [expanded, activeTab, selectedPod])

  const loadPods = async () => {
    setPodsLoading(true)
    try {
      const data = await getPods(job.id)
      const list = Array.isArray(data) ? data : []
      setPods(list)
      if (list.length > 0 && (!selectedPod || !list.find(p => p.name === selectedPod))) {
        setSelectedPod(list[0].name)
      }
    } catch (e) {
      console.error(e)
    } finally {
      setPodsLoading(false)
    }
  }

  const loadPodLogs = async (name) => {
    setPodLogsLoading(true)
    try {
      const data = await getPodLogs(job.id, name)
      setPodLogs(data.logs || '')
    } catch (e) {
      setPodLogs(`Failed to fetch logs: ${e.message}`)
    } finally {
      setPodLogsLoading(false)
    }
  }

  const handleScale = async () => {
    setScaling(true)
    try {
      await scaleDeploy(job.id, replicas)
      // refresh pods list
      await loadPods()
    } catch (e) {
      alert(`Scaling failed: ${e.message}`)
    } finally {
      setScaling(false)
    }
  }

  const handleDelete = async () => {
    if (!window.confirm(`Are you sure you want to completely delete "${job.app_name}"? This will purge all running Kubernetes resources.`)) return
    try {
      await deleteDeploy(job.id)
      if (onDeployDeleted) onDeployDeleted(job.id)
    } catch (e) {
      alert(`Deletion failed: ${e.message}`)
    }
  }

  return (
    <div className="glass-panel" style={{
      borderRadius: 'var(--radius)',
      padding: '18px 24px',
      border: '1px solid var(--border)',
      transition: 'all 0.25s ease',
    }}
    onMouseEnter={e => e.currentTarget.style.borderColor = 'var(--border-hover)'}
    onMouseLeave={e => e.currentTarget.style.borderColor = 'var(--border)'}
    >
      {/* Header row */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 14, flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1, minWidth: 0 }}>
          <span style={{
            width: 36, height: 36, borderRadius: 'var(--radius-sm)',
            background: 'linear-gradient(135deg, var(--accent) 0%, #4facfe 100%)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 15, fontWeight: 700, flexShrink: 0, textShadow: '0 1px 3px rgba(0,0,0,0.2)',
          }}>
            {job.app_name?.[0]?.toUpperCase() ?? '?'}
          </span>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontWeight: 700, fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {job.app_name}
            </div>
            <div style={{ color: 'var(--muted)', fontSize: 11 }}>
              {job.id?.slice(0, 8)} · {ts}
            </div>
          </div>
        </div>

        <StatusBadge status={job.status} />

        {job.url && job.status === 'live' && (
          <a href={job.url} target="_blank" rel="noreferrer" style={{
            fontSize: 12, color: 'var(--green)', display: 'inline-flex', alignItems: 'center', gap: 4,
            fontWeight: 600, padding: '4px 10px', background: 'var(--green-glow)', borderRadius: 6,
          }}>
            Open App ↗
          </a>
        )}

        <button
          onClick={() => setExpanded(v => !v)}
          className="premium-btn"
          style={{
            background: 'transparent', border: '1px solid var(--border)',
            borderRadius: 'var(--radius-sm)', padding: '5px 12px', color: 'var(--text)',
            fontSize: 12, cursor: 'pointer', boxShadow: 'none',
          }}
          onMouseEnter={e => e.currentTarget.style.background = 'rgba(255,255,255,0.02)'}
          onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
        >
          {expanded ? 'Collapse' : 'Manage'}
        </button>
      </div>

      {/* Expanded panel */}
      {expanded && (
        <div style={{ marginTop: 18, borderTop: '1px solid var(--border)', paddingTop: 16, animation: 'slideDown 0.2s ease-out' }}>
          {/* Tabs bar */}
          <div style={{ display: 'flex', borderBottom: '1px solid var(--border)', marginBottom: 16 }}>
            {[
              { id: 'build', label: '🛠 Build Logs' },
              { id: 'pods', label: '🩺 Pod Health' },
              { id: 'runtime', label: '💻 Runtime Terminal' },
              { id: 'settings', label: '⚙ Settings' },
            ].map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                style={{
                  background: 'transparent', border: 'none', borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
                  padding: '8px 16px', fontSize: 12.5, fontWeight: activeTab === tab.id ? 700 : 500,
                  color: activeTab === tab.id ? 'var(--text)' : 'var(--muted)', cursor: 'pointer',
                  borderRadius: 0,
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* Tab contents */}
          {activeTab === 'build' && (
            <div>
              <LogViewer logs={job.logs} />
            </div>
          )}

          {activeTab === 'pods' && (
            <div>
              {podsLoading ? (
                <div style={{ fontSize: 12, color: 'var(--muted)', textAlign: 'center', padding: '24px 0' }}>Loading pod status...</div>
              ) : pods.length === 0 ? (
                <div style={{ fontSize: 12, color: 'var(--muted)', textAlign: 'center', padding: '24px 0' }}>No active pods. Application may be scaled to 0 or deploying.</div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {pods.map(pod => (
                    <div key={pod.name} style={{
                      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                      background: 'rgba(255,255,255,0.01)', border: '1px solid var(--border)',
                      borderRadius: 'var(--radius-sm)', padding: '10px 14px', fontSize: 12.5,
                    }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <span style={{
                          width: 8, height: 8, borderRadius: '50%',
                          background: pod.status === 'Running' ? 'var(--green)' : pod.status === 'Pending' ? 'var(--amber)' : 'var(--red)',
                          boxShadow: pod.status === 'Running' ? '0 0 8px var(--green)' : 'none',
                        }} />
                        <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>{pod.name}</span>
                      </div>
                      <div style={{ display: 'flex', gap: 16, color: 'var(--muted)', fontSize: 12 }}>
                        <span>Status: <b style={{ color: 'var(--text)' }}>{pod.status}</b></span>
                        <span>Restarts: <b style={{ color: pod.restarts > 0 ? 'var(--red)' : 'var(--text)' }}>{pod.restarts}</b></span>
                        <span>Age: <b style={{ color: 'var(--text)' }}>{pod.age}</b></span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'runtime' && (
            <div>
              <div style={{ display: 'flex', gap: 10, marginBottom: 12, alignItems: 'center' }}>
                <span style={{ fontSize: 12, color: 'var(--muted)' }}>Select Pod:</span>
                <select
                  value={selectedPod}
                  onChange={e => setSelectedPod(e.target.value)}
                  style={{
                    background: 'var(--bg3)', border: '1px solid var(--border)',
                    borderRadius: 6, color: 'var(--text)', padding: '4px 8px', fontSize: 12,
                    outline: 'none',
                  }}
                >
                  {pods.map(p => (
                    <option key={p.name} value={p.name}>{p.name}</option>
                  ))}
                </select>
                <button
                  onClick={() => loadPodLogs(selectedPod)}
                  disabled={podLogsLoading || !selectedPod}
                  style={{
                    background: 'transparent', border: '1px solid var(--border)',
                    borderRadius: 6, padding: '3px 10px', fontSize: 11, color: 'var(--text)',
                    cursor: 'pointer',
                  }}
                  onMouseEnter={e => e.currentTarget.style.background = 'rgba(255,255,255,0.02)'}
                  onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                >
                  {podLogsLoading ? 'Refreshing...' : '🔄 Refresh'}
                </button>
              </div>

              {podLogsLoading ? (
                <div style={{ fontSize: 12, color: 'var(--muted)', textAlign: 'center', padding: '24px 0' }}>Streaming runtime console...</div>
              ) : !selectedPod ? (
                <div style={{ fontSize: 12, color: 'var(--muted)', textAlign: 'center', padding: '24px 0' }}>No pods available to inspect logs.</div>
              ) : (
                <div className="terminal" style={{ maxHeight: 250, fontSize: 11 }}>
                  {podLogs ? (
                    <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{podLogs}</pre>
                  ) : (
                    <div style={{ color: 'var(--muted)', fontStyle: 'italic' }}>Pod started, waiting for console stdout...</div>
                  )}
                </div>
              )}
            </div>
          )}

          {activeTab === 'settings' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
              {/* Scaling */}
              <div style={{
                background: 'rgba(255,255,255,0.01)', border: '1px solid var(--border)',
                borderRadius: 'var(--radius)', padding: '16px 20px',
              }}>
                <div style={{ fontWeight: 700, fontSize: 14, marginBottom: 12 }}>Replicas Scaling</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                  <input
                    type="range" min="0" max="10"
                    value={replicas}
                    onChange={e => setReplicas(parseInt(e.target.value))}
                    disabled={scaling}
                    style={{ flex: 1, accentColor: 'var(--accent)' }}
                  />
                  <span style={{ fontSize: 14, fontWeight: 700, width: 24, textScale: 'center' }}>{replicas}</span>
                  <button
                    onClick={handleScale}
                    disabled={scaling || replicas === job.replicas}
                    className="premium-btn"
                    style={{ padding: '6px 16px', fontSize: 12, borderRadius: 6 }}
                  >
                    {scaling ? 'Scaling...' : 'Scale'}
                  </button>
                </div>
                <div style={{ fontSize: 11.5, color: 'var(--muted)', marginTop: 8 }}>
                  Scale the number of running container instances. Setting replicas to 0 will shut down the app.
                </div>
              </div>

              {/* Env vars list */}
              {job.env_vars && Object.keys(job.env_vars).length > 0 && (
                <div style={{
                  background: 'rgba(255,255,255,0.01)', border: '1px solid var(--border)',
                  borderRadius: 'var(--radius)', padding: '16px 20px',
                }}>
                  <div style={{ fontWeight: 700, fontSize: 14, marginBottom: 12 }}>Environment Config</div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                    {Object.entries(job.env_vars).map(([k, v]) => (
                      <div key={k} style={{ display: 'flex', fontSize: 12, borderBottom: '1px solid rgba(255,255,255,0.02)', paddingBottom: 4 }}>
                        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: 'var(--accent)', width: 120 }}>{k}</span>
                        <span style={{ fontFamily: 'monospace', color: 'var(--text)', wordBreak: 'break-all' }}>{v}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Danger zone */}
              <div style={{
                border: '1px solid rgba(239, 68, 68, 0.15)', background: 'rgba(239, 68, 68, 0.02)',
                borderRadius: 'var(--radius)', padding: '16px 20px',
              }}>
                <div style={{ fontWeight: 700, fontSize: 14, color: 'var(--red)', marginBottom: 6 }}>Danger Zone</div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 10 }}>
                  <span style={{ fontSize: 11.5, color: 'var(--muted)' }}>
                    Completely remove this application deployment and release all Kubernetes resources.
                  </span>
                  <button
                    onClick={handleDelete}
                    style={{
                      background: 'var(--red)', color: '#fff', border: 'none',
                      borderRadius: 6, padding: '7px 16px', fontSize: 12, fontWeight: 700,
                      boxShadow: '0 4px 10px rgba(239, 68, 68, 0.2)',
                    }}
                    onMouseEnter={e => e.currentTarget.style.background = '#dc2626'}
                    onMouseLeave={e => e.currentTarget.style.background = 'var(--red)'}
                  >
                    Delete App
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
