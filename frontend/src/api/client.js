const BASE = '/deploy'

export async function submitDeploy(appName, zipFile, envVars = {}) {
  const form = new FormData()
  form.append('app_name', appName)
  form.append('source', zipFile)
  if (Object.keys(envVars).length > 0) {
    form.append('env_vars', JSON.stringify(envVars))
  }
  const res = await fetch(BASE, { method: 'POST', body: form })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Deploy failed')
  return data
}

export async function getStatus(id) {
  const res = await fetch(`${BASE}/${id}`)
  if (!res.ok) throw new Error('Not found')
  return res.json()
}

export async function listDeploys() {
  const res = await fetch(BASE)
  if (!res.ok) throw new Error('Failed to list deployments')
  return res.json()
}

export async function deleteDeploy(id) {
  const res = await fetch(`${BASE}/${id}`, { method: 'DELETE' })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Failed to delete deployment')
  return data
}

export async function scaleDeploy(id, replicas) {
  const res = await fetch(`${BASE}/${id}/scale?replicas=${replicas}`, { method: 'POST' })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Failed to scale deployment')
  return data
}

export async function getPods(id) {
  const res = await fetch(`${BASE}/${id}/pods`)
  if (!res.ok) throw new Error('Failed to fetch pods')
  return res.json()
}

export async function getPodLogs(id, podName) {
  const res = await fetch(`${BASE}/${id}/pods/${podName}/logs`)
  if (!res.ok) throw new Error('Failed to fetch pod logs')
  return res.json()
}
