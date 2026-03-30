import React, { useEffect, useRef } from 'react'

export default function LogViewer({ logs = [], maxHeight = 220 }) {
  const ref = useRef(null)

  useEffect(() => {
    if (ref.current) {
      ref.current.scrollTop = ref.current.scrollHeight
    }
  }, [logs])

  if (!logs.length) return null

  return (
    <div ref={ref} style={{
      background: '#09090c',
      border: '1px solid rgba(255,255,255,0.06)',
      borderRadius: 8,
      padding: '12px 14px',
      maxHeight,
      overflowY: 'auto',
      marginTop: 12,
    }}>
      {logs.map((line, i) => {
        const isSuccess = line.startsWith('✓')
        const isError   = line.startsWith('✗')
        const isStep    = line.startsWith('▶')
        const color = isSuccess ? '#22c77a' : isError ? '#e05454' : isStep ? '#7c6ff7' : '#a0a0b0'
        return (
          <div key={i} style={{ color, fontFamily: 'monospace', fontSize: 11.5, lineHeight: '1.7' }}>
            {line}
          </div>
        )
      })}
    </div>
  )
}
