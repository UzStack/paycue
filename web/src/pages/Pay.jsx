import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api'

function formatAmount(n) {
  if (n === null || n === undefined) return ''
  return Number(n).toLocaleString('uz-UZ')
}

// Karta raqamini 4 talab guruhlaydi (to'lovchi o'qishi uchun).
function formatCard(num, last4) {
  if (!num) return last4 ? `**** ${last4}` : '—'
  return String(num).replace(/\s/g, '').replace(/(.{4})/g, '$1 ').trim()
}

function pad(n) {
  return String(n).padStart(2, '0')
}

function CopyButton({ value, label }) {
  const [copied, setCopied] = useState(false)
  async function copy() {
    try {
      await navigator.clipboard.writeText(String(value))
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {}
  }
  return (
    <button
      onClick={copy}
      className="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium border border-zinc-700 text-zinc-300 hover:text-white hover:border-zinc-500 transition-colors shrink-0"
      title={label}
    >
      {copied ? (
        <>
          <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
          Nusxalandi
        </>
      ) : (
        <>
          <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
          Nusxalash
        </>
      )}
    </button>
  )
}

const STATES = {
  active: {
    label: 'To\'lov kutilmoqda',
    color: 'text-sky-400',
    ring: 'border-sky-500/30',
    dot: 'bg-sky-400',
  },
  confirmed: {
    label: 'To\'lov qabul qilindi',
    color: 'text-emerald-400',
    ring: 'border-emerald-500/30',
    dot: 'bg-emerald-400',
  },
  cancelled: {
    label: 'Muddati o\'tdi',
    color: 'text-red-400',
    ring: 'border-red-500/30',
    dot: 'bg-red-400',
  },
  expired: {
    label: 'Muddati tugamoqda',
    color: 'text-amber-400',
    ring: 'border-amber-500/30',
    dot: 'bg-amber-400',
  },
}

export default function Pay() {
  const { id } = useParams()
  const [info, setInfo] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [now, setNow] = useState(() => Date.now())

  const fetchInfo = useCallback(async () => {
    try {
      const data = await api.payInfo(id)
      setInfo(data)
      setError('')
      return data
    } catch (err) {
      setError(err.message)
      return null
    } finally {
      setLoading(false)
    }
  }, [id])

  // Dastlabki yuklash
  useEffect(() => { fetchInfo() }, [fetchInfo])

  // To'lov hali ochiqmi (active yoki worker hali yopmagan expired) — soat va polling shunga bog'liq.
  const open = !!info && (info.state === 'active' || info.state === 'expired')

  // Har soniya soatni yangilab turamiz (countdown uchun) — faqat ochiq holatda, terminal holatda to'xtaydi.
  useEffect(() => {
    if (!open) return
    const t = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(t)
  }, [open])

  // Holatni davriy so'rab turamiz — interval faqat holat o'zgarganda qayta yaratiladi (har poll'da emas).
  useEffect(() => {
    if (!open) return
    const id = setInterval(fetchInfo, 4000)
    return () => clearInterval(id)
  }, [open, fetchInfo])

  if (loading) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
        <div className="w-full max-w-md">
          <div className="h-72 bg-zinc-900 border border-zinc-800 rounded-2xl animate-pulse" />
        </div>
      </div>
    )
  }

  if (error || !info) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center px-4 text-center">
        <div className="max-w-md">
          <div className="w-14 h-14 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center text-red-400 mx-auto mb-5">
            <svg width="26" height="26" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" /></svg>
          </div>
          <h1 className="text-xl font-bold text-zinc-100 mb-2">To'lov topilmadi</h1>
          <p className="text-zinc-400 text-sm">Havola noto'g'ri yoki muddati o'tgan bo'lishi mumkin.</p>
        </div>
      </div>
    )
  }

  const s = STATES[info.state] || STATES.active
  const expiresAt = info.expires_at ? new Date(info.expires_at).getTime() : 0
  const remainingMs = Math.max(0, expiresAt - now)
  const totalMs = (info.timeout_mins || 30) * 60 * 1000
  const progress = totalMs ? Math.max(0, Math.min(100, (remainingMs / totalMs) * 100)) : 0
  const mins = Math.floor(remainingMs / 60000)
  const secs = Math.floor((remainingMs % 60000) / 1000)
  const isActive = info.state === 'active'
  const isConfirmed = info.state === 'confirmed'
  const isCancelled = info.state === 'cancelled'
  const timeUp = remainingMs <= 0
  // 'expired' (worker hali yopmagan) yoki vaqt tugagan active — "tekshirilmoqda" holati.
  const isVerifying = info.state === 'expired' || (isActive && timeUp)

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 flex flex-col items-center justify-center px-4 py-10">
      {/* Brand */}
      <div className="flex items-center gap-2 mb-6">
        <span className="text-sky-400 font-bold text-lg tracking-tight">Paycue</span>
      </div>

      <div className="w-full max-w-md">
        <div className={`bg-zinc-900 border ${s.ring} rounded-2xl overflow-hidden shadow-xl shadow-black/30`}>
          {/* Status header */}
          <div className="px-6 pt-6 pb-5 border-b border-zinc-800 text-center">
            {isConfirmed ? (
              <div className="w-16 h-16 rounded-full bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 mx-auto mb-4">
                <svg width="32" height="32" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
              </div>
            ) : isCancelled ? (
              <div className="w-16 h-16 rounded-full bg-red-500/10 border border-red-500/30 flex items-center justify-center text-red-400 mx-auto mb-4">
                <svg width="30" height="30" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
              </div>
            ) : (
              <div className="relative w-16 h-16 mx-auto mb-4">
                <span className="absolute inset-0 rounded-full bg-sky-500/20 animate-ping" />
                <span className="relative w-16 h-16 rounded-full bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400">
                  <svg width="28" height="28" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                </span>
              </div>
            )}
            <div className={`inline-flex items-center gap-2 text-sm font-semibold ${isVerifying ? 'text-amber-400' : s.color}`}>
              {!isConfirmed && !isCancelled && <span className={`w-1.5 h-1.5 rounded-full ${s.dot} animate-pulse`} />}
              {isVerifying ? 'To\'lov tekshirilmoqda...' : s.label}
            </div>
          </div>

          {/* Amount */}
          <div className="px-6 py-6 text-center">
            <p className="text-zinc-500 text-xs uppercase tracking-wide mb-2">To'lov summasi</p>
            <div className="flex items-center justify-center gap-3">
              <span className="text-4xl font-bold font-mono text-zinc-50">{formatAmount(info.amount)}</span>
              <span className="text-zinc-500 text-lg">UZS</span>
            </div>
            {isActive && (
              <div className="mt-3 inline-flex items-start gap-1.5 text-xs text-amber-400/90 bg-amber-500/5 border border-amber-500/15 rounded-md px-3 py-1.5 text-left">
                <svg width="14" height="14" className="mt-px shrink-0" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" /></svg>
                Aynan shu summani o'tkazing — summa noyob, shuning uchun to'lov avtomatik aniqlanadi.
              </div>
            )}
          </div>

          {/* Card to pay */}
          <div className="px-6 pb-2">
            <div className="bg-zinc-950/60 border border-zinc-800 rounded-xl p-4">
              <div className="flex items-center justify-between gap-3 mb-3">
                <div className="min-w-0">
                  <p className="text-zinc-500 text-xs mb-1">Karta raqami</p>
                  <p className="text-zinc-100 font-mono text-base tracking-wide break-all">{formatCard(info.card_number, info.card_last4)}</p>
                </div>
                {info.card_number && <CopyButton value={String(info.card_number).replace(/\s/g, '')} label="Karta raqamini nusxalash" />}
              </div>
              {info.card_owner && (
                <div className="flex items-center justify-between border-t border-zinc-800 pt-3">
                  <span className="text-zinc-500 text-xs">Karta egasi</span>
                  <span className="text-zinc-300 text-sm">{info.card_owner}</span>
                </div>
              )}
            </div>
          </div>

          {/* Countdown / footer */}
          <div className="px-6 py-5">
            {isActive && !timeUp && (
              <>
                <div className="flex items-center justify-between text-xs text-zinc-500 mb-2">
                  <span>Qolgan vaqt</span>
                  <span className="font-mono text-zinc-300 text-sm">{pad(mins)}:{pad(secs)}</span>
                </div>
                <div className="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-sky-500 rounded-full transition-all duration-1000 ease-linear"
                    style={{ width: `${progress}%` }}
                  />
                </div>
                <p className="text-zinc-500 text-xs text-center mt-4">
                  To'lov tasdiqlangach bu sahifa avtomatik yangilanadi.
                </p>
              </>
            )}
            {isVerifying && (
              <p className="text-zinc-400 text-sm text-center">Muddat tugadi, to'lov holati tekshirilmoqda...</p>
            )}
            {isConfirmed && (
              <p className="text-emerald-400/90 text-sm text-center">Rahmat! To'lov muvaffaqiyatli qabul qilindi.</p>
            )}
            {isCancelled && (
              <p className="text-zinc-400 text-sm text-center">Ushbu to'lov muddati o'tgan. Yangi to'lov uchun qaytadan urinib ko'ring.</p>
            )}
          </div>
        </div>

        <p className="text-center text-zinc-600 text-xs mt-5">
          Xavfsiz to'lov · <span className="text-zinc-500">Paycue</span>
        </p>
      </div>
    </div>
  )
}
