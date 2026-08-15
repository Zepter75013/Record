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

export async function fetchLatestReportMetadata() {
  const response = await fetch(`${API_BASE_URL}/reports/latest`, {
    method: 'GET',
    headers: authHeaders({ Accept: 'application/json' }),
  })

  if (response.status === 204) {
    return null
  }

  if (!response.ok) {
    throw new Error('Impossible de charger le dernier état généré.')
  }

  return parseJson(response)
}

export async function uploadGeneratedReport(blob, criteria, format) {
  const formData = new FormData()
  formData.append('file', blob, `rapport-disques-manager.${format}`)
  formData.append('criteria', criteria)
  formData.append('format', format)

  const response = await fetch(`${API_BASE_URL}/reports`, {
    method: 'POST',
    headers: authHeaders({ Accept: 'application/json' }),
    body: formData,
  })

  if (!response.ok) {
    const message = await extractErrorMessage(response, 'Impossible d’enregistrer le rapport généré.')
    throw new Error(message)
  }

  return parseJson(response)
}

// Le PDF s'ouvre dans un nouvel onglet (le navigateur sait l'afficher tel
// quel) ; les autres formats (XLSX/DOCX/CSV) n'ont pas de rendu inline
// possible dans un onglet — on déclenche donc un téléchargement classique
// avec le bon nom de fichier/extension.
export async function viewLatestReportFile(format) {
  const response = await fetch(`${API_BASE_URL}/reports/latest/file`, {
    method: 'GET',
    headers: authHeaders(),
  })

  if (!response.ok) {
    const message = await extractErrorMessage(response, 'Impossible d’ouvrir le rapport.')
    throw new Error(message)
  }

  const blob = await response.blob()
  const url = URL.createObjectURL(blob)

  if (format === 'pdf') {
    const win = window.open(url, '_blank', 'noopener')
    if (!win) window.location.href = url
  } else {
    const link = document.createElement('a')
    link.href = url
    link.download = `rapport-disques-manager.${format || 'pdf'}`
    document.body.appendChild(link)
    link.click()
    link.remove()
  }

  setTimeout(() => URL.revokeObjectURL(url), 60000)
}
