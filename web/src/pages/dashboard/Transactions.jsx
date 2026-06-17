import { useEffect, useMemo, useState } from 'react'
import { api } from '../../api'

const PAGE_SIZE = 15

function formatAmount(n) {
  if (n === null || n === undefined) return ''
  return Number(n).toLocaleString('uz-UZ')
}

function formatDateTime(str) {
  if (!str) return ''
  try {
    return new Date(str).toLocaleString('uz-UZ', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit',
    })
  } catch { return str }
}

// To'liq karta raqamini 4 talab bo'lib ko'rsatadi (yashirilmaydi — apidan keladi).
function formatCard(num, last4) {
  if (!num) return last4 ? `*${last4}` : '—'
  const clean = String(num).replace(/\s/g, '')
  return clean.replace(/(.{4})/g, '$1 ').trim()
}

const STATES = {
  active: { label: 'Aktiv', cls: 'text-sky-400 bg-sky-500/10 border-sky-500/20' },
  confirmed: { label: 'Tasdiqlangan', cls: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20' },
  cancelled: { label: 'Bekor qilingan', cls: 'text-red-400 bg-red-500/10 border-red-500/20' },
  expired: { label: 'Muddati o\'tgan', cls: 'text-amber-400 bg-amber-500/10 border-amber-500/20' },
}

function StateBadge({ state }) {
  const s = STATES[state] || { label: state || '—', cls: 'text-zinc-400 bg-zinc-800 border-zinc-700' }
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${s.cls}`}>
      {s.label}
    </span>
  )
}

// Webhook yetkazib berish holati + urinishlar soni (transaction yopilgandan keyin).
function WebhookChip({ tx }) {
  // Hali ochiq (active/expired) — webhook yuborilmagan.
  if (tx.state === 'active' || tx.state === 'expired') {
    return <span className="text-xs text-zinc-600">webhook: kutilmoqda</span>
  }
  const attempts = tx.webhook_attempts || 0
  if (attempts === 0) {
    return <span className="text-xs text-zinc-500">webhook: sozlanmagan</span>
  }
  const cls = tx.webhook_status
    ? 'text-emerald-400'
    : 'text-red-400'
  return (
    <span className={`inline-flex items-center gap-1 text-xs ${cls}`}>
      {tx.webhook_status ? (
        <svg width="11" height="11" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
      ) : (
        <svg width="11" height="11" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
      )}
      webhook {tx.webhook_status ? 'yuborildi' : 'xato'} · {attempts}x
    </span>
  )
}

const FILTERS = [
  { key: 'all', label: 'Hammasi' },
  { key: 'active', label: 'Aktiv' },
  { key: 'confirmed', label: 'Tasdiqlangan' },
  { key: 'cancelled', label: 'Bekor qilingan' },
  { key: 'expired', label: 'Muddati o\'tgan' },
]

export default function Transactions() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('all')
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState(() => new Set())
  const [busy, setBusy] = useState(false)

  function load() {
    setLoading(true)
    api.transactionList()
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  // Filter + qidiruv
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    return items.filter((t) => {
      if (filter !== 'all' && t.state !== filter) return false
      if (!q) return true
      return (
        String(t.amount).includes(q) ||
        (t.card_number || '').toLowerCase().includes(q) ||
        (t.card_last4 || '').includes(q) ||
        (t.card_owner || '').toLowerCase().includes(q) ||
        (t.transaction_id || '').toLowerCase().includes(q)
      )
    })
  }, [items, filter, query])

  // Filter yoki qidiruv o'zgarsa birinchi sahifaga qaytamiz
  useEffect(() => { setPage(1); setSelected(new Set()) }, [filter, query])

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE))
  const pageItems = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  // Holatlar bo'yicha sanoq (filter tablari uchun)
  const counts = useMemo(() => {
    const c = { all: items.length, active: 0, confirmed: 0, cancelled: 0, expired: 0 }
    for (const t of items) if (c[t.state] !== undefined) c[t.state]++
    return c
  }, [items])

  function toggle(id) {
    setSelected((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  function toggleAllOnPage() {
    setSelected((prev) => {
      const next = new Set(prev)
      const allSelected = pageItems.every((t) => next.has(t.id))
      pageItems.forEach((t) => allSelected ? next.delete(t.id) : next.add(t.id))
      return next
    })
  }

  async function handleDelete(id) {
    if (!window.confirm('Ushbu tranzaksiyani o\'chirasizmi?')) return
    setBusy(true)
    try {
      await api.transactionDelete(id)
      setItems((prev) => prev.filter((t) => t.id !== id))
      setSelected((prev) => { const n = new Set(prev); n.delete(id); return n })
    } catch (err) {
      alert('Xato: ' + err.message)
    } finally {
      setBusy(false)
    }
  }

  async function handleBulkDelete() {
    const ids = [...selected]
    if (!ids.length) return
    if (!window.confirm(`${ids.length} ta tranzaksiyani o'chirasizmi?`)) return
    setBusy(true)
    try {
      await Promise.all(ids.map((id) => api.transactionDelete(id).catch(() => null)))
      setItems((prev) => prev.filter((t) => !selected.has(t.id)))
      setSelected(new Set())
    } catch (err) {
      alert('Xato: ' + err.message)
    } finally {
      setBusy(false)
    }
  }

  const allOnPageSelected = pageItems.length > 0 && pageItems.every((t) => selected.has(t.id))

  return (
    <div>
      <div className="flex items-center justify-between mb-6 gap-3">
        <div>
          <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Tranzaksiyalar</h1>
          <p className="text-zinc-400 text-sm mt-1">Yaratilgan to'lovlar va ularning holati</p>
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

      {/* Qidiruv + bulk action */}
      <div className="flex flex-wrap items-center gap-3 mb-4">
        <div className="relative flex-1 min-w-[200px]">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
          </svg>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Summa, karta raqami yoki egasi bo'yicha qidirish..."
            className="w-full pl-9 pr-3 py-2 bg-zinc-900 border border-zinc-800 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
          />
        </div>
        {selected.size > 0 && (
          <button
            onClick={handleBulkDelete}
            disabled={busy}
            className="flex items-center gap-2 px-3 py-2 bg-red-500/10 hover:bg-red-500/20 disabled:opacity-50 text-red-400 border border-red-500/30 rounded-md text-sm font-medium transition-colors shrink-0"
          >
            <svg width="15" height="15" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
            {selected.size} ta o'chirish
          </button>
        )}
      </div>

      {error && (
        <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2 mb-4">
          {error}
        </p>
      )}

      {/* Ro'yxat */}
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
              <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
            </svg>
          </div>
          <p className="text-zinc-400 text-sm">
            {items.length === 0 ? 'Hali tranzaksiya yaratilmagan' : 'Filtr bo\'yicha tranzaksiya topilmadi'}
          </p>
        </div>
      ) : (
        <>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            {/* Sarlavha (desktop) */}
            <div className="hidden md:flex items-center gap-3 px-5 py-2.5 border-b border-zinc-800 text-xs font-medium text-zinc-500">
              <input
                type="checkbox"
                checked={allOnPageSelected}
                onChange={toggleAllOnPage}
                className="w-4 h-4 rounded border-zinc-600 bg-zinc-800 accent-sky-500 cursor-pointer"
              />
              <span className="w-28">Summa</span>
              <span className="flex-1">Karta</span>
              <span className="w-36">Holat / Webhook</span>
              <span className="w-40">Sana</span>
              <span className="w-8" />
            </div>

            <div className="divide-y divide-zinc-800">
              {pageItems.map((t) => (
                <div key={t.id} className="flex flex-col md:flex-row md:items-center gap-2 md:gap-3 px-5 py-3.5">
                  <div className="flex items-center gap-3 md:contents">
                    <input
                      type="checkbox"
                      checked={selected.has(t.id)}
                      onChange={() => toggle(t.id)}
                      className="w-4 h-4 rounded border-zinc-600 bg-zinc-800 accent-sky-500 cursor-pointer shrink-0"
                    />
                    <span className="md:w-28 text-zinc-100 text-sm font-semibold font-mono">
                      {formatAmount(t.amount)} <span className="text-zinc-500 font-normal text-xs">UZS</span>
                    </span>
                  </div>
                  <div className="flex-1 min-w-0 pl-7 md:pl-0">
                    <p className="text-zinc-300 text-sm font-mono truncate">{formatCard(t.card_number, t.card_last4)}</p>
                    <p className="text-zinc-500 text-xs truncate">{t.card_owner || '—'}</p>
                  </div>
                  <div className="md:w-36 pl-7 md:pl-0 flex flex-col gap-1 items-start">
                    <StateBadge state={t.state} />
                    <WebhookChip tx={t} />
                  </div>
                  <span className="md:w-40 pl-7 md:pl-0 text-zinc-500 text-xs">{formatDateTime(t.created_at)}</span>
                  <button
                    onClick={() => handleDelete(t.id)}
                    disabled={busy}
                    title="O'chirish"
                    className="md:w-8 self-end md:self-auto text-zinc-500 hover:text-red-400 disabled:opacity-50 transition-colors"
                  >
                    <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Pagination */}
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
