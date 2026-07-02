import React, { useState, useRef } from 'react'
import { submitDeploy } from '../api/client.js'

export default function DeployForm({ onDeployStarted }) {
  const [appName, setAppName] = useState('')
  const [file, setFile]       = useState(null)
  const [error, setError]     = useState('')
  const [loading, setLoading] = useState(false)
  const [envVars, setEnvVars] = useState([{ key: '', value: '' }])
  const [showEnv, setShowEnv] = useState(false)
  const fileRef = useRef(null)

  const handleAppName = (v) => {
    setAppName(v.toLowerCase().replace(/[^a-z0-9-]/g, ''))
  }

  const handleAddEnv = () => {
    setEnvVars(prev => [...prev, { key: '', value: '' }])
  }

  const handleRemoveEnv = (index) => {
    setEnvVars(prev => prev.filter((_, i) => i !== index))
  }

  const handleEnvChange = (index, field, val) => {
    setEnvVars(prev => prev.map((item, i) => i === index ? { ...item, [field]: val } : item))
  }

  const handleSubmit = async () => {
    setError('')
    if (!appName || appName.length < 3) { setError('App name must be at least 3 characters'); return }
    if (!file) { setError('Please select a zip file'); return }

    const envObj = {}
    envVars.forEach(item => {
      if (item.key.trim()) {
        envObj[item.key.trim().toUpperCase()] = item.value.trim()
      }
    })

    setLoading(true)
    try {
      const data = await submitDeploy(appName, file, envObj)
      onDeployStarted(data.deploy_id)
      setAppName('')
      setFile(null)
      setEnvVars([{ key: '', value: '' }])
      setShowEnv(false)
      if (fileRef.current) fileRef.current.value = ''
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="glass-panel" style={{
      borderRadius: 'var(--radius-lg)',
      padding: '24px',
      marginBottom: 32,
      border: '1px solid var(--border)',
    }}>
      <div style={{ fontWeight: 700, fontSize: 16, marginBottom: 20, display: 'flex', alignItems: 'center', gap: 10 }}>
        <span style={{ fontSize: 18 }}>🚀</span> New Deployment
      </div>

      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap', marginBottom: 20 }}>
        {/* App name */}
        <div style={{ flex: '1 1 200px' }}>
          <label style={{ display: 'block', fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginBottom: 6, letterSpacing: '0.05em', textTransform: 'uppercase' }}>
            App Name
          </label>
          <input
            value={appName}
            onChange={e => handleAppName(e.target.value)}
            placeholder="my-cool-app"
            disabled={loading}
            className="premium-input"
          />
        </div>

        {/* File pick */}
        <div style={{ flex: '2 1 280px' }}>
          <label style={{ display: 'block', fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginBottom: 6, letterSpacing: '0.05em', textTransform: 'uppercase' }}>
            Source ZIP
          </label>
          <div
            style={{
              width: '100%', padding: '10.5px 14px',
              background: 'rgba(255, 255, 255, 0.02)', border: '1px dashed var(--border)',
              borderRadius: 'var(--radius)', color: file ? 'var(--text)' : 'var(--muted)',
              cursor: 'pointer', fontSize: 13, display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              transition: 'all 0.2s',
            }}
            onClick={() => fileRef.current?.click()}
            onMouseEnter={e => e.currentTarget.style.borderColor = 'var(--accent)'}
            onMouseLeave={e => e.currentTarget.style.borderColor = 'var(--border)'}
          >
            <span>{file ? `📦 ${file.name}` : 'Select source .zip archive...'}</span>
            {file && <span style={{ fontSize: 11, color: 'var(--muted)' }}>{(file.size / 1024 / 1024).toFixed(2)} MB</span>}
          </div>
          <input
            ref={fileRef}
            type="file"
            accept=".zip"
            style={{ display: 'none' }}
            onChange={e => setFile(e.target.files[0] || null)}
            disabled={loading}
          />
        </div>
      </div>

      {/* Env Toggle */}
      <div style={{ marginBottom: 20 }}>
        <button
          onClick={() => setShowEnv(v => !v)}
          style={{
            background: 'transparent', border: 'none', color: 'var(--muted)',
            fontSize: 13, fontWeight: 600, display: 'flex', alignItems: 'center', gap: 6,
            padding: 0,
          }}
        >
          <span>{showEnv ? '▼' : '▶'}</span>
          <span>Environment Variables ({envVars.filter(e => e.key).length})</span>
        </button>

        {showEnv && (
          <div style={{
            marginTop: 12, padding: '16px', background: 'rgba(255,255,255,0.01)',
            borderRadius: 'var(--radius)', border: '1px solid var(--border)',
            display: 'flex', flexDirection: 'column', gap: 8,
            animation: 'slideDown 0.2s ease-out',
          }}>
            {envVars.map((item, idx) => (
              <div key={idx} style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <input
                  placeholder="KEY (e.g. PORT)"
                  value={item.key}
                  onChange={e => handleEnvChange(idx, 'key', e.target.value)}
                  disabled={loading}
                  className="premium-input"
                  style={{ textTransform: 'uppercase', flex: 1 }}
                />
                <input
                  placeholder="VALUE"
                  value={item.value}
                  onChange={e => handleEnvChange(idx, 'value', e.target.value)}
                  disabled={loading}
                  className="premium-input"
                  style={{ flex: 1 }}
                />
                <button
                  onClick={() => handleRemoveEnv(idx)}
                  disabled={loading}
                  style={{
                    background: 'rgba(239,68,68,0.1)', color: 'var(--red)',
                    border: 'none', borderRadius: 'var(--radius-sm)',
                    padding: '8px 12px', fontSize: 12, fontWeight: 600,
                  }}
                  onMouseEnter={e => e.currentTarget.style.background = 'var(--red)'}
                  onMouseLeave={e => e.currentTarget.style.background = 'rgba(239,68,68,0.1)'}
                >
                  ✕
                </button>
              </div>
            ))}
            <button
              onClick={handleAddEnv}
              disabled={loading}
              style={{
                alignSelf: 'flex-start', background: 'transparent',
                border: '1px dashed var(--border)', borderRadius: 'var(--radius-sm)',
                color: 'var(--accent)', padding: '6px 12px', fontSize: 12,
                fontWeight: 600, marginTop: 4, cursor: 'pointer',
              }}
              onMouseEnter={e => e.currentTarget.style.borderColor = 'var(--accent)'}
              onMouseLeave={e => e.currentTarget.style.borderColor = 'var(--border)'}
            >
              + Add Variable
            </button>
          </div>
        )}
      </div>

      {error && (
        <div style={{ color: 'var(--red)', fontSize: 12, marginBottom: 16, padding: '10px 14px', background: 'var(--red-glow)', border: '1px solid rgba(239,68,68,0.15)', borderRadius: 'var(--radius)' }}>
          ⚠️ {error}
        </div>
      )}

      <button
        onClick={handleSubmit}
        disabled={loading || !appName || !file}
        className="premium-btn"
        style={{ width: '100%' }}
      >
        {loading ? '⏳ Deploying...' : 'Deploy App →'}
      </button>
    </div>
  )
}
