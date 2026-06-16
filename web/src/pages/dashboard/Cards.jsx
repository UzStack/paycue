import { useEffect, useState } from 'react'
import { api } from '../../api'

function formatDate(str) {
  if (!str) return ''
  try {
    return new Date(str).toLocaleDateString('uz-UZ', { year: 'numeric', month: 'short', day: 'numeric' })
  } catch { return str }
}

function maskNumber(num) {
  if (!num) return ''
  const clean = num.replace(/\s/g, '')
  if (clean.length < 4) return clean
  return clean.slice(0, 4) + ' **** **** ' + clean.slice(-4)
}

export default function Cards() {
  const [cards, setCards] = useState([])
  const [accounts, setAccounts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ telegram_account_id: '', number: '', owner_name: '' })
  const [formLoading, setFormLoading] = useState(false)
  const [formError, setFormError] = useState('')

  function load() {
    setLoading(true)
    Promise.all([api.cardList(), api.telegramList()])
      .then(([c, tg]) => {
        setCards(Array.isArray(c) ? c : [])
        setAccounts(Array.isArray(tg) ? tg.filter(a => a.status === 'active') : [])
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  async function handleSubmit(e) {
    e.preventDefault()
    setFormError('')
    if (!form.telegram_account_id) { setFormError('Telegram akkaunt tanlang'); return }
    if (!form.number.trim()) { setFormError('Karta raqami kiriting'); return }
    if (!form.owner_name.trim()) { setFormError('Egasi ismini kiriting'); return }
    setFormLoading(true)
    try {
      await api.cardCreate({
        telegram_account_id: Number(form.telegram_account_id),
        number: form.number.replace(/\s/g, ''),
        owner_name: form.owner_name.trim(),
      })
      setShowForm(false)
      setForm({ telegram_account_id: '', number: '', owner_name: '' })
      load()
    } catch (err) {
      setFormError(err.message)
    } finally {
      setFormLoading(false)
    }
  }

  function getAccountPhone(id) {
    const acc = accounts.find(a => String(a.id) === String(id)) ||
      (accounts.length ? null : null)
    return acc ? acc.phone : id
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Kartalar</h1>
          <p className="text-zinc-400 text-sm mt-1">Ulangan Humo kartalar</p>
        </div>
        <button
          onClick={() => { setShowForm(!showForm); setFormError('') }}
          className="flex items-center gap-2 px-4 py-2 bg-sky-500 hover:bg-sky-400 text-white font-medium rounded-md text-sm transition-colors"
        >
          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          Karta qo'shish
        </button>
      </div>

      {/* Qo'shish formasi */}
      {showForm && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-zinc-100">Yangi karta</h2>
            <button
              onClick={() => setShowForm(false)}
              className="text-zinc-500 hover:text-zinc-300 transition-colors"
            >
              <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">
                Telegram akkaunt <span className="text-red-400">*</span>
              </label>
              <select
                value={form.telegram_account_id}
                onChange={(e) => setForm({ ...form, telegram_account_id: e.target.value })}
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={formLoading}
              >
                <option value="">Akkaunt tanlang</option>
                {accounts.map((acc) => (
                  <option key={acc.id} value={acc.id}>{acc.phone}{acc.username ? ` (@${acc.username})` : ''}</option>
                ))}
              </select>
              {accounts.length === 0 && (
                <p className="text-xs text-amber-400 mt-1">Faol Telegram akkaunt yo'q. Avval akkaunt ulang.</p>
              )}
            </div>

            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">
                Karta raqami <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={form.number}
                onChange={(e) => setForm({ ...form, number: e.target.value })}
                placeholder="9860 1234 5678 9012"
                maxLength={19}
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm font-mono focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={formLoading}
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">
                Karta egasi <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={form.owner_name}
                onChange={(e) => setForm({ ...form, owner_name: e.target.value })}
                placeholder="ALISHER TOSHMATOV"
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={formLoading}
              />
            </div>

            {formError && (
              <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2">
                {formError}
              </p>
            )}

            <button
              type="submit"
              disabled={formLoading}
              className="w-full py-2 bg-sky-500 hover:bg-sky-400 disabled:bg-sky-500/50 text-white font-medium rounded-md text-sm transition-colors"
            >
              {formLoading ? 'Qo\'shilmoqda...' : 'Karta qo\'shish'}
            </button>
          </form>
        </div>
      )}

      {/* Ro'yxat */}
      {loading ? (
        <div className="space-y-3">
          {[1, 2].map((i) => (
            <div key={i} className="h-20 bg-zinc-900 border border-zinc-800 rounded-lg animate-pulse" />
          ))}
        </div>
      ) : cards.length === 0 ? (
        <div className="text-center py-12 bg-zinc-900 border border-zinc-800 rounded-lg">
          <div className="w-10 h-10 rounded-md bg-zinc-800 flex items-center justify-center text-zinc-500 mx-auto mb-3">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" />
            </svg>
          </div>
          <p className="text-zinc-400 text-sm">Hech qanday karta qo'shilmagan</p>
          <p className="text-zinc-600 text-xs mt-1">Yuqoridagi "Karta qo'shish" tugmasini bosing</p>
        </div>
      ) : (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
          <div className="divide-y divide-zinc-800">
            {cards.map((card) => (
              <div key={card.id} className="flex items-center justify-between px-5 py-4">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-8 h-8 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 shrink-0">
                    <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" />
                    </svg>
                  </div>
                  <div className="min-w-0">
                    <p className="text-zinc-100 text-sm font-medium font-mono">
                      {maskNumber(card.number)}
                    </p>
                    <p className="text-zinc-500 text-xs">{card.owner_name}</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 shrink-0 text-right">
                  <div className="hidden sm:block">
                    <p className="text-zinc-400 text-xs">
                      {getAccountPhone(card.telegram_account_id)}
                    </p>
                    <p className="text-zinc-600 text-xs">{formatDate(card.created_at)}</p>
                  </div>
                  <span className="text-xs text-zinc-500 font-mono bg-zinc-800 px-2 py-0.5 rounded">
                    *{card.last4}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
