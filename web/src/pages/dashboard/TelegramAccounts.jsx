import { useEffect, useState } from 'react'
import { api } from '../../api'

function StatusBadge({ status }) {
  if (status === 'active') {
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
        <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
        Faol
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-amber-500/10 border border-amber-500/20 text-amber-400 text-xs font-medium">
      <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
      Kutilmoqda
    </span>
  )
}

function formatDate(str) {
  if (!str) return ''
  try {
    return new Date(str).toLocaleDateString('uz-UZ', { year: 'numeric', month: 'short', day: 'numeric' })
  } catch {
    return str
  }
}

// Ulash oqimi holatlari: idle | send-code | verify | need-password | done
const FLOW_IDLE = 'idle'
const FLOW_SEND = 'send-code'
const FLOW_VERIFY = 'verify'
const FLOW_PASSWORD = 'need-password'
const FLOW_DONE = 'done'

export default function TelegramAccounts() {
  const [accounts, setAccounts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)

  // Oqim holati
  const [flowState, setFlowState] = useState(FLOW_IDLE)
  const [phone, setPhone] = useState('')
  const [tgAccountId, setTgAccountId] = useState(null)
  const [code, setCode] = useState('')
  const [twoFaPass, setTwoFaPass] = useState('')
  const [flowLoading, setFlowLoading] = useState(false)
  const [flowError, setFlowError] = useState('')
  const [flowMsg, setFlowMsg] = useState('')

  function loadAccounts() {
    setLoading(true)
    api.telegramList()
      .then((data) => setAccounts(Array.isArray(data) ? data : []))
      .catch(() => setAccounts([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadAccounts() }, [])

  function openForm() {
    setShowForm(true)
    setFlowState(FLOW_SEND)
    setPhone('')
    setCode('')
    setTwoFaPass('')
    setTgAccountId(null)
    setFlowError('')
    setFlowMsg('')
  }

  function closeForm() {
    setShowForm(false)
    setFlowState(FLOW_IDLE)
  }

  async function handleSendCode(e) {
    e.preventDefault()
    setFlowError('')
    if (!phone.trim()) { setFlowError('Telefon raqam kiriting'); return }
    setFlowLoading(true)
    try {
      const data = await api.telegramSendCode({ phone: phone.trim() })
      setTgAccountId(data.telegram_account_id)
      setFlowMsg(data.message || 'Kod yuborildi')
      setFlowState(FLOW_VERIFY)
    } catch (err) {
      setFlowError(err.message)
    } finally {
      setFlowLoading(false)
    }
  }

  async function handleVerify(e) {
    e.preventDefault()
    setFlowError('')
    if (!code.trim()) { setFlowError('Kodni kiriting'); return }
    setFlowLoading(true)
    try {
      const body = { telegram_account_id: tgAccountId, code: code.trim() }
      if (flowState === FLOW_PASSWORD && twoFaPass) body.password = twoFaPass
      const data = await api.telegramVerify(body)
      if (data.need_password) {
        setFlowState(FLOW_PASSWORD)
        setFlowError('')
      } else {
        setFlowState(FLOW_DONE)
        loadAccounts()
      }
    } catch (err) {
      setFlowError(err.message)
    } finally {
      setFlowLoading(false)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Telegram akkauntlar</h1>
          <p className="text-zinc-400 text-sm mt-1">Ulangan Telegram akkauntlaringiz</p>
        </div>
        <button
          onClick={openForm}
          className="flex items-center gap-2 px-4 py-2 bg-sky-500 hover:bg-sky-400 text-white font-medium rounded-md text-sm transition-colors"
        >
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          Ulash
        </button>
      </div>

      {/* Ulash formasi */}
      {showForm && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-zinc-100">
              {flowState === FLOW_SEND && 'Telegram akkaunt ulash'}
              {flowState === FLOW_VERIFY && 'Tasdiqlash kodi'}
              {flowState === FLOW_PASSWORD && '2FA paroli'}
              {flowState === FLOW_DONE && 'Akkaunt ulandi'}
            </h2>
            <button onClick={closeForm} className="text-zinc-500 hover:text-zinc-300 transition-colors">
              <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {flowState === FLOW_SEND && (
            <form onSubmit={handleSendCode} className="flex gap-3">
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+998901234567"
                className="flex-1 px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={flowLoading}
              />
              <button
                type="submit"
                disabled={flowLoading}
                className="px-4 py-2 bg-sky-500 hover:bg-sky-400 disabled:bg-sky-500/50 text-white font-medium rounded-md text-sm transition-colors shrink-0"
              >
                {flowLoading ? 'Yuborilmoqda...' : 'Kod yuborish'}
              </button>
            </form>
          )}

          {(flowState === FLOW_VERIFY || flowState === FLOW_PASSWORD) && (
            <form onSubmit={handleVerify} className="space-y-3">
              {flowMsg && (
                <p className="text-xs text-zinc-400 bg-zinc-800 px-3 py-2 rounded-md">{flowMsg}</p>
              )}
              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-1">
                  Tasdiqlash kodi
                </label>
                <input
                  type="text"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="123456"
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                  disabled={flowLoading}
                />
              </div>
              {flowState === FLOW_PASSWORD && (
                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">
                    2FA paroli (Telegram Cloud Password)
                  </label>
                  <input
                    type="password"
                    value={twoFaPass}
                    onChange={(e) => setTwoFaPass(e.target.value)}
                    placeholder="Cloud parolingiz"
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                    disabled={flowLoading}
                  />
                </div>
              )}
              <button
                type="submit"
                disabled={flowLoading}
                className="w-full py-2 bg-sky-500 hover:bg-sky-400 disabled:bg-sky-500/50 text-white font-medium rounded-md text-sm transition-colors"
              >
                {flowLoading ? 'Tekshirilmoqda...' : 'Tasdiqlash'}
              </button>
            </form>
          )}

          {flowState === FLOW_DONE && (
            <div className="flex items-center gap-3 text-emerald-400">
              <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm font-medium">Akkaunt muvaffaqiyatli ulandi</span>
            </div>
          )}

          {flowError && (
            <p className="mt-3 text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2">
              {flowError}
            </p>
          )}
        </div>
      )}

      {/* Ro'yxat */}
      {loading ? (
        <div className="space-y-3">
          {[1, 2].map((i) => (
            <div key={i} className="h-16 bg-zinc-900 border border-zinc-800 rounded-lg animate-pulse" />
          ))}
        </div>
      ) : accounts.length === 0 ? (
        <div className="text-center py-12 bg-zinc-900 border border-zinc-800 rounded-lg">
          <div className="w-10 h-10 rounded-md bg-zinc-800 flex items-center justify-center text-zinc-500 mx-auto mb-3">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
            </svg>
          </div>
          <p className="text-zinc-400 text-sm">Hech qanday akkaunt ulangan emas</p>
          <p className="text-zinc-600 text-xs mt-1">Yuqoridagi "Ulash" tugmasini bosing</p>
        </div>
      ) : (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
          <div className="divide-y divide-zinc-800">
            {accounts.map((acc) => (
              <div key={acc.id} className="flex items-center justify-between px-5 py-4">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-8 h-8 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 shrink-0">
                    <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
                    </svg>
                  </div>
                  <div className="min-w-0">
                    <p className="text-zinc-100 text-sm font-medium truncate">{acc.phone}</p>
                    {acc.username && (
                      <p className="text-zinc-500 text-xs">@{acc.username}</p>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-4 shrink-0">
                  <span className="text-zinc-600 text-xs hidden sm:block">{formatDate(acc.created_at)}</span>
                  <StatusBadge status={acc.status} />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
