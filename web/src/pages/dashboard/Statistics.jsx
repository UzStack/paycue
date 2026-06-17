import { useEffect, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { api } from '../../api'

function n(v) {
  return Number(v || 0).toLocaleString('uz-UZ')
}

function StatCard({ label, value, hint, icon }) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
      <div className="flex items-start justify-between mb-3">
        <span className="text-zinc-400 text-sm font-medium">{label}</span>
        {icon && (
          <div className="w-8 h-8 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 shrink-0">
            {icon}
          </div>
        )}
      </div>
      <span className="text-2xl font-bold text-zinc-100">{value}</span>
      {hint && <p className="text-zinc-500 text-xs mt-1">{hint}</p>}
    </div>
  )
}

export default function Statistics() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function load() {
    setLoading(true)
    api.stats()
      .then((d) => setData(d))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  if (loading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
          <div key={i} className="h-24 bg-zinc-900 border border-zinc-800 rounded-lg animate-pulse" />
        ))}
      </div>
    )
  }

  // Backend statistikani o'chirib qo'ygan bo'lsa — sahifa ochilmaydi.
  if (!error && data && data.enabled === false) {
    return <Navigate to="/dashboard" replace />
  }

  const txTotal = data?.transactions || 0
  const confirmed = data?.transactions_confirmed || 0
  const confirmRate = txTotal ? Math.round((confirmed / txTotal) * 100) : 0
  const versions = data?.versions || {}
  const versionList = Object.entries(versions).sort((a, b) => b[1] - a[1])

  return (
    <div>
      <div className="flex items-center justify-between mb-6 gap-3">
        <div>
          <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Statistika</h1>
          <p className="text-zinc-400 text-sm mt-1">Barcha Paycue instance'lari bo'yicha anonim jamlanma</p>
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

      {error && (
        <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2 mb-4">
          {error}
        </p>
      )}

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
        <StatCard
          label="Instance'lar"
          value={n(data?.instances)}
          hint="Hisobot yuborgan serverlar"
          icon={<svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M21.75 17.25v-.228a4.5 4.5 0 00-.12-1.03l-2.268-9.64a3.375 3.375 0 00-3.285-2.602H7.923a3.375 3.375 0 00-3.285 2.602l-2.268 9.64a4.5 4.5 0 00-.12 1.03v.228m19.5 0a3 3 0 01-3 3H5.25a3 3 0 01-3-3m19.5 0a3 3 0 00-3-3H5.25a3 3 0 00-3 3m16.5 0h.008v.008h-.008v-.008zm-3 0h.008v.008h-.008v-.008z" /></svg>}
        />
        <StatCard label="Foydalanuvchilar" value={n(data?.users)} icon={<svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>} />
        <StatCard label="Telegram akkauntlar" value={n(data?.telegram_accounts)} icon={<svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" /></svg>} />
        <StatCard label="Kartalar" value={n(data?.cards)} icon={<svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" /></svg>} />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <StatCard label="Jami tranzaksiya" value={n(txTotal)} />
        <StatCard label="Tasdiqlangan" value={n(confirmed)} hint={`${confirmRate}% konversiya`} />
        <StatCard label="Bekor qilingan" value={n(data?.transactions_cancelled)} />
        <StatCard label="Webhook loglar" value={n(data?.webhook_logs)} />
      </div>

      {/* Versiyalar taqsimoti */}
      {versionList.length > 0 && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
          <h2 className="text-sm font-semibold text-zinc-100 mb-4">Versiyalar bo'yicha instance'lar</h2>
          <div className="space-y-3">
            {versionList.map(([ver, cnt]) => {
              const pct = data.instances ? Math.round((cnt / data.instances) * 100) : 0
              return (
                <div key={ver}>
                  <div className="flex items-center justify-between text-xs mb-1">
                    <span className="text-zinc-300 font-mono">{ver}</span>
                    <span className="text-zinc-500">{cnt} ta · {pct}%</span>
                  </div>
                  <div className="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                    <div className="h-full bg-sky-500 rounded-full" style={{ width: `${pct}%` }} />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      <p className="text-zinc-600 text-xs mt-4">
        Ma'lumotlar anonim: faqat sanoqlar yig'iladi, hech qanday maxfiy ma'lumot (token, telefon, karta) saqlanmaydi.
      </p>
    </div>
  )
}
