import { getAuthToken } from '@/utils/authToken'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

function authHeaders(extra = {}) {
  const token = getAuthToken()
  return {
    ...extra,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

async function parseJson(response) {
  const contentType = response.headers.get('content-type') || ''

  if (!contentType.includes('application/json')) {
    throw new Error('Réponse API invalide.')
  }

  return response.json()
}

async function extractErrorMessage(response, fallbackMessage) {
  let message = fallbackMessage

  try {
    const contentType = response.headers.get('content-type') || ''

    if (contentType.includes('application/json')) {
      const errorData = await response.json()
      if (errorData?.message) {
        message = errorData.message
      }
    } else {
      const text = (await response.text()).trim()
      if (text) message = text
    }
  } catch {
    // on garde le message par défaut
  }

  return message
}

export async function fetchTracks(discId) {
  const response = await fetch(`${API_BASE_URL}/discs/${encodeURIComponent(discId)}/tracks`, {
    method: 'GET',
    headers: authHeaders({ Accept: 'application/json' }),
  })

  if (!response.ok) {
    const message = await extractErrorMessage(response, 'Impossible de charger les pistes.')
    throw new Error(message)
  }

  return parseJson(response)
}

// force=true remplace une tracklist déjà existante par une nouvelle
// recherche Discogs (au lieu de ne rien faire si des pistes existent déjà).
export async function fetchTracklistOnDemand(discId, force = false) {
  const url = `${API_BASE_URL}/discs/${encodeURIComponent(discId)}/tracks/fetch${force ? '?force=true' : ''}`
  const response = await fetch(url, {
    method: 'POST',
    headers: authHeaders({ Accept: 'application/json' }),
  })

  if (!response.ok) {
    const message = await extractErrorMessage(response, 'Impossible de récupérer les pistes sur Discogs.')
    throw new Error(message)
  }

  return parseJson(response)
}

export async function updateTracks(discId, tracks) {
  const response = await fetch(`${API_BASE_URL}/discs/${encodeURIComponent(discId)}/tracks`, {
    method: 'PUT',
    headers: authHeaders({ 'Content-Type': 'application/json', Accept: 'application/json' }),
    body: JSON.stringify(
      tracks.map((t) => ({ position: t.position || '', title: t.title, duration: t.duration || '' }))
    ),
  })

  if (!response.ok) {
    const message = await extractErrorMessage(response, 'Impossible d’enregistrer les pistes.')
    throw new Error(message)
  }

  return parseJson(response)
}
