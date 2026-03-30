import React, { useState, useRef } from 'react'
import { submitDeploy } from '../api/client.js'

export default function DeployForm({ onDeployStarted }) {
  const [appName, setAppName] = useState('')
  const [file, setFile]       = useState(null)
  const [error, setError]     = useState('')
  const [loading, setLoading] = useState(false)
  const fileRef = useRef(null)

  const handleAppName = (v) => {
    setAppName(v.toLowerCase().replace(/[^a-z0-9-]/g, ''))
  }

  const handleSubmit = async () => {
    setError('')
    if (!appName || appName.length < 3) { setError('App name must be at least 3 characters'); return }
    if (!file) { setError('Please select a zip file'); return }

    setLoading(true)
    try {
      const data = await submitDeploy(appName, file)
      onDeployStarted(data.deploy_id)
      setAppName('')
      setFile(null)
      if (fileRef.current) fileRef.current.value = ''
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      background: 'var(--bg2)',
      border: '1px solid var(--border)',
      borderRadius: 'var(--radius-lg)',
      padding: '20px 24px',
      marginBottom: 24,
    }}>
      <div style={{ fontWeight: 600, fontSize: 15, marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
        <span>🚀</span> New Deployment
      </div>

      <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', marginBottom: 12 }}>
        {/* App name */}
        <div style={{ flex: '1 1 180px' }}>
          <label style={{ display: 'block', fontSize: 11, color: 'var(--muted)', marginBottom: 5, letterSpacing: '0.05em', textTransform: 'uppercase' }}>
            App Name
          </label>
          <input
            value={appName}
            onChange={e => handleAppName(e.target.value)}
            placeholder="my-app"
            disabled={loading}
            style={{
              width: '100%', padding: '8px 12px',
              background: 'var(--bg3)', border: '1px solid var(--border2)',
              borderRadius: 8, color: 'var(--text)', outline: 'none',
            }}
          />
        </div>

        {/* File pick */}
        <div style={{ flex: '2 1 220px' }}>
          <label style={{ display: 'block', fontSize: 11, color: 'var(--muted)', marginBottom: 5, letterSpacing: '0.05em', textTransform: 'uppercase' }}>
            Source ZIP
          </label>
          <div
            style={{
              width: '100%', padding: '8px 12px',
              background: 'var(--bg3)', border: '1px dashed var(--border2)',
              borderRadius: 8, color: file ? 'var(--text)' : 'var(--muted)',
              cursor: 'pointer', fontSize: 13,
            }}
            onClick={() => fileRef.current?.click()}
          >
            {file ? `📦 ${file.name}` : 'Click to select .zip file...'}
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

      {error && (
        <div style={{ color: 'var(--red)', fontSize: 12, marginBottom: 10, padding: '8px 12px', background: 'rgba(224,84,84,0.08)', borderRadius: 6 }}>
          {error}
        </div>
      )}

      <button
        onClick={handleSubmit}
        disabled={loading || !appName || !file}
        style={{
          background: loading || !appName || !file ? 'var(--bg3)' : 'var(--accent)',
          color: loading || !appName || !file ? 'var(--muted)' : '#fff',
          border: 'none', borderRadius: 8,
          padding: '9px 22px', fontWeight: 600,
          fontSize: 14, transition: 'all 0.15s',
          cursor: loading || !appName || !file ? 'not-allowed' : 'pointer',
        }}
      >
        {loading ? '⏳ Deploying...' : 'Deploy →'}
      </button>
    </div>
  )
}
