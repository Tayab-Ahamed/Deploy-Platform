const BASE = '/deploy'

export async function submitDeploy(appName, zipFile) {
  const form = new FormData()
  form.append('app_name', appName)
  form.append('source', zipFile)
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
