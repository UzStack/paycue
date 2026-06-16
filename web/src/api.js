const BASE = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080'

function getToken() {
  return localStorage.getItem('paycue_token')
}

async function request(path, options = {}) {
  const token = getToken()
  const headers = { 'Content-Type': 'application/json', ...options.headers }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${BASE}${path}`, { ...options, headers })
  const json = await res.json()
  if (!json.status) throw new Error(json.data?.detail || 'Xato yuz berdi')
  return json.data
}

export const api = {
  register: (body) => request('/api/register', { method: 'POST', body: JSON.stringify(body) }),
  login: (body) => request('/api/login', { method: 'POST', body: JSON.stringify(body) }),
  getWebhook: () => request('/api/webhook'),
  setWebhook: (body) => request('/api/webhook', { method: 'POST', body: JSON.stringify(body) }),
  telegramSendCode: (body) => request('/api/telegram/send-code', { method: 'POST', body: JSON.stringify(body) }),
  telegramVerify: (body) => request('/api/telegram/verify', { method: 'POST', body: JSON.stringify(body) }),
  telegramList: () => request('/api/telegram'),
  cardCreate: (body) => request('/api/cards', { method: 'POST', body: JSON.stringify(body) }),
  cardList: () => request('/api/cards'),
  transactionCreate: (body) => request('/api/transactions', { method: 'POST', body: JSON.stringify(body) }),
}

export { getToken }
