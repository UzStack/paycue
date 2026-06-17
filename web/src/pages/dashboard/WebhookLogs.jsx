import { useEffect, useMemo, useState } from 'react'
import { api } from '../../api'

const PAGE_SIZE = 20

function formatDateTime(str) {
  if (!str) return ''
  try {
    return new Date(str).toLocaleString('uz-UZ', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit',
    })
  } catch { return str }
}

function formatAmount(n) {
  if (n === null || n === undefined) return ''
  return Number(n).toLocaleString('uz-UZ')
}

const ACTION_LABEL = {
  confirm: { label: 'To\'lov tushdi', cls: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20' },
  cancel: { label: 'Bekor qilindi', cls: 'text-amber-400 bg-amber-500/10 border-amber-500/20' },
}

function ActionBadge({ action }) {
  const a = ACTION_LABEL[action] || { label: action || '—', cls: 'text-zinc-400 bg-zinc-800 border-zinc-700' }
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${a.cls}`}>
      {a.label}
    </span>
  )
}

const FILTERS = [
  { key: 'all', label: 'Hammasi' },
  { key: 'success', label: 'Muvaffaqiyatli' },
  { key: 'failed', label: 'Xato' },
]

export default function WebhookLogs() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('all')
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)

  function load() {
    setLoading(true)
    api.webhookLogs()
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    return items.filter((l) => {
      if (filter === 'success' && !l.success) return false
      if (filter === 'failed' && l.success) return false
      if (!q) return true
      return (
        String(l.amount).includes(q) ||
        (l.transaction_id || '').toLowerCase().includes(q) ||
        (l.url || '').toLowerCase().includes(q) ||
        String(l.status_code).includes(q)
      )
    })
  }, [items, filter, query])

  useEffect(() => { setPage(1) }, [filter, query])

  const counts = useMemo(() => {
    let success = 0, failed = 0
    for (const l of items) l.success ? success++ : failed++
    return { all: items.length, success, failed }
  }, [items])

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE))
  const pageItems = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div>
      <div className="flex items-center justify-between mb-6 gap-3">
        <div>
          <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Webhook loglar</h1>
          <p className="text-zinc-400 text-sm mt-1">Barcha webhook yetkazib berishlar va ularning natijasi</p>
        </div>
        <button
          onClick={load}
          className="flex items-center gap-2 px-3 py-2 border border-zinc-700 hover:border-zinc-500 text-zinc-300 hover:text-zinc-100 rounded-md text-sm transition-colors shrink-0"
        >
          <svg width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992V4.356M3.985 14.652H-.007v4.992M4.94 19.644a8.25 8.25 0 0013.803-3.7M19.06 4.356a8.25 8.25 0 00-13.803 3.7" />
          </svg>
          Yangilash
        </button>
      </div>

      {/* Filter tablari */}
      <div className="flex flex-wrap gap-2 mb-4">
        {FILTERS.map((f) => (
          <button
            key={f.key}
            onClick={() => setFilter(f.key)}
            className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors border ${
              filter === f.key
                ? 'bg-sky-500/10 text-sky-400 border-sky-500/30'
                : 'text-zinc-400 border-zinc-800 hover:text-zinc-200 hover:border-zinc-700'
            }`}
          >
            {f.label}
            <span className="ml-1.5 text-xs opacity-60">{counts[f.key] ?? 0}</span>
          </button>
        ))}
      </div>

      {/* Qidiruv */}
      <div className="relative mb-4">
        <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
        </svg>
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Summa, transaction ID, URL yoki status kod bo'yicha..."
          className="w-full pl-9 pr-3 py-2 bg-zinc-900 border border-zinc-800 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
        />
      </div>

      {error && (
        <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2 mb-4">
          {error}
        </p>
      )}

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 bg-zinc-900 border border-zinc-800 rounded-lg animate-pulse" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="text-center py-12 bg-zinc-900 border border-zinc-800 rounded-lg">
          <div className="w-10 h-10 rounded-md bg-zinc-800 flex items-center justify-center text-zinc-500 mx-auto mb-3">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
            </svg>
          </div>
          <p className="text-zinc-400 text-sm">
            {items.length === 0 ? 'Hali webhook yuborilmagan' : 'Filtr bo\'yicha log topilmadi'}
          </p>
          {items.length === 0 && (
            <p className="text-zinc-600 text-xs mt-1">Tranzaksiya yopilganda webhook yuboriladi va shu yerda ko'rinadi</p>
          )}
        </div>
      ) : (
        <>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <div className="hidden md:flex items-center gap-3 px-5 py-2.5 border-b border-zinc-800 text-xs font-medium text-zinc-500">
              <span className="w-24">Natija</span>
              <span className="w-28">Summa</span>
              <span className="w-32">Hodisa</span>
              <span className="flex-1">Transaction</span>
              <span className="w-20 text-center">Kod</span>
              <span className="w-16 text-center">Urinish</span>
              <span className="w-40">Sana</span>
            </div>

            <div className="divide-y divide-zinc-800">
              {pageItems.map((l) => (
                <div key={l.id} className="flex flex-col md:flex-row md:items-center gap-2 md:gap-3 px-5 py-3.5">
                  {/* Natija */}
                  <div className="md:w-24">
                    {l.success ? (
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-400">
                        <svg width="12" height="12" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
                        Yuborildi
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-red-400">
                        <svg width="12" height="12" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                        Xato
                      </span>
                    )}
                  </div>
                  <span className="md:w-28 text-zinc-100 text-sm font-semibold font-mono">
                    {formatAmount(l.amount)} <span className="text-zinc-500 font-normal text-xs">UZS</span>
                  </span>
                  <div className="md:w-32"><ActionBadge action={l.action} /></div>
                  <span className="flex-1 min-w-0 text-zinc-400 text-xs font-mono truncate" title={l.transaction_id}>
                    {l.transaction_id || '—'}
                  </span>
                  <span className={`md:w-20 md:text-center text-xs font-mono ${l.status_code >= 200 && l.status_code < 300 ? 'text-emerald-400' : l.status_code ? 'text-red-400' : 'text-zinc-600'}`}>
                    {l.status_code || '—'}
                  </span>
                  <span className="md:w-16 md:text-center text-zinc-300 text-xs font-mono">{l.attempts}x</span>
                  <span className="md:w-40 text-zinc-500 text-xs">{formatDateTime(l.created_at)}</span>
                  {l.error && !l.success && (
                    <p className="md:hidden text-xs text-red-400/80 break-words">{l.error}</p>
                  )}
                </div>
              ))}
            </div>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-zinc-500 text-xs">
                {filtered.length} ta natija · {page}/{totalPages} sahifa
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-3 py-1.5 border border-zinc-800 rounded-md text-sm text-zinc-300 hover:border-zinc-600 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  Oldingi
                </button>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-3 py-1.5 border border-zinc-800 rounded-md text-sm text-zinc-300 hover:border-zinc-600 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  Keyingi
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
