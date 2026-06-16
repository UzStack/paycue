import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../api'

function StatCard({ label, value, loading, icon }) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
      <div className="flex items-start justify-between mb-3">
        <span className="text-zinc-400 text-sm font-medium">{label}</span>
        <div className="w-8 h-8 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400">
          {icon}
        </div>
      </div>
      {loading ? (
        <div className="h-7 w-12 bg-zinc-800 rounded animate-pulse" />
      ) : (
        <span className="text-2xl font-bold text-zinc-100">{value}</span>
      )}
    </div>
  )
}

export default function Overview() {
  const [tgCount, setTgCount] = useState(0)
  const [cardCount, setCardCount] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([api.telegramList(), api.cardList()])
      .then(([tg, cards]) => {
        setTgCount(Array.isArray(tg) ? tg.length : 0)
        setCardCount(Array.isArray(cards) ? cards.length : 0)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Bosh sahifa</h1>
        <p className="text-zinc-400 text-sm mt-1">Hisobingiz holati</p>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-8">
        <StatCard
          label="Telegram akkauntlar"
          value={tgCount}
          loading={loading}
          icon={
            <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
            </svg>
          }
        />
        <StatCard
          label="Kartalar"
          value={cardCount}
          loading={loading}
          icon={
            <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" />
            </svg>
          }
        />
      </div>

      <div>
        <h2 className="text-sm font-medium text-zinc-400 mb-3 uppercase tracking-wider">Tezkor amallar</h2>
        <div className="grid sm:grid-cols-2 gap-3">
          <Link
            to="/dashboard/telegram"
            className="flex items-center gap-3 px-4 py-3.5 bg-zinc-900 border border-zinc-800 rounded-lg hover:border-zinc-700 hover:bg-zinc-800/70 transition-colors group"
          >
            <div className="w-9 h-9 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 group-hover:bg-sky-500/15 transition-colors shrink-0">
              <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
              </svg>
            </div>
            <div>
              <p className="text-zinc-100 text-sm font-medium">Telegram ulash</p>
              <p className="text-zinc-500 text-xs mt-0.5">Yangi akkaunt qo'shish</p>
            </div>
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" className="text-zinc-600 ml-auto">
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
          </Link>

          <Link
            to="/dashboard/cards"
            className="flex items-center gap-3 px-4 py-3.5 bg-zinc-900 border border-zinc-800 rounded-lg hover:border-zinc-700 hover:bg-zinc-800/70 transition-colors group"
          >
            <div className="w-9 h-9 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 group-hover:bg-sky-500/15 transition-colors shrink-0">
              <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" />
              </svg>
            </div>
            <div>
              <p className="text-zinc-100 text-sm font-medium">Karta qo'shish</p>
              <p className="text-zinc-500 text-xs mt-0.5">Humo karta ulash</p>
            </div>
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" className="text-zinc-600 ml-auto">
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
          </Link>

          <Link
            to="/dashboard/webhook"
            className="flex items-center gap-3 px-4 py-3.5 bg-zinc-900 border border-zinc-800 rounded-lg hover:border-zinc-700 hover:bg-zinc-800/70 transition-colors group"
          >
            <div className="w-9 h-9 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 group-hover:bg-sky-500/15 transition-colors shrink-0">
              <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
              </svg>
            </div>
            <div>
              <p className="text-zinc-100 text-sm font-medium">Webhook sozlash</p>
              <p className="text-zinc-500 text-xs mt-0.5">Callback URL kiriting</p>
            </div>
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" className="text-zinc-600 ml-auto">
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
          </Link>

          <Link
            to="/dashboard/transaction"
            className="flex items-center gap-3 px-4 py-3.5 bg-zinc-900 border border-zinc-800 rounded-lg hover:border-zinc-700 hover:bg-zinc-800/70 transition-colors group"
          >
            <div className="w-9 h-9 rounded-md bg-sky-500/10 flex items-center justify-center text-sky-400 group-hover:bg-sky-500/15 transition-colors shrink-0">
              <svg width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
              </svg>
            </div>
            <div>
              <p className="text-zinc-100 text-sm font-medium">Tranzaksiya yaratish</p>
              <p className="text-zinc-500 text-xs mt-0.5">To'lov yarating</p>
            </div>
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" className="text-zinc-600 ml-auto">
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
          </Link>
        </div>
      </div>
    </div>
  )
}
