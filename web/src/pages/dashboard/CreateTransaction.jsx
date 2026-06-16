import { useEffect, useState } from 'react'
import { api } from '../../api'

function formatAmount(n) {
  if (!n) return ''
  return Number(n).toLocaleString('uz-UZ')
}

export default function CreateTransaction() {
  const [cards, setCards] = useState([])
  const [cardsLoading, setCardsLoading] = useState(true)
  const [amount, setAmount] = useState('')
  const [cardId, setCardId] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState(null)

  useEffect(() => {
    api.cardList()
      .then((data) => setCards(Array.isArray(data) ? data : []))
      .catch(() => setCards([]))
      .finally(() => setCardsLoading(false))
  }, [])

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setResult(null)
    const amt = Number(amount)
    if (!amt || amt <= 0) { setError('To\'g\'ri summa kiriting'); return }

    setLoading(true)
    try {
      const body = { amount: amt }
      if (cardId) body.card_id = Number(cardId)
      const data = await api.transactionCreate(body)
      setResult(data)
      setAmount('')
      setCardId('')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Tranzaksiya yaratish</h1>
        <p className="text-zinc-400 text-sm mt-1">To'lov yarating va natijani ko'ring</p>
      </div>

      <div className="max-w-md">
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-4">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">
                Summa (so'm) <span className="text-red-400">*</span>
              </label>
              <div className="relative">
                <input
                  type="number"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  placeholder="50000"
                  min="1"
                  className="w-full px-3 py-2.5 pr-12 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm font-mono focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                  disabled={loading}
                />
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 text-xs">UZS</span>
              </div>
            </div>

            <div>
              <label className="block text-xs font-medium text-zinc-400 mb-1.5">
                Karta <span className="text-zinc-500 font-normal">(ixtiyoriy)</span>
              </label>
              {cardsLoading ? (
                <div className="h-9 bg-zinc-800 rounded-md animate-pulse" />
              ) : (
                <select
                  value={cardId}
                  onChange={(e) => setCardId(e.target.value)}
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                  disabled={loading}
                >
                  <option value="">Avtomatik tanlash</option>
                  {cards.map((card) => (
                    <option key={card.id} value={card.id}>
                      *{card.last4} - {card.owner_name}
                    </option>
                  ))}
                </select>
              )}
              <p className="text-xs text-zinc-500 mt-1">
                Tanlanmasa, eng kam yukli karta avtomatik ishlatiladi
              </p>
            </div>

            {error && (
              <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 bg-sky-500 hover:bg-sky-400 disabled:bg-sky-500/50 disabled:cursor-not-allowed text-white font-medium rounded-md text-sm transition-colors flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <svg width="14" height="14" className="animate-spin" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Yaratilmoqda...
                </>
              ) : (
                <>
                  <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                  </svg>
                  To'lov yaratish
                </>
              )}
            </button>
          </form>
        </div>

        {/* Natija */}
        {result && (
          <div className="bg-emerald-500/5 border border-emerald-500/20 rounded-lg p-5">
            <div className="flex items-center gap-2 mb-4">
              <div className="w-8 h-8 rounded-md bg-emerald-500/10 flex items-center justify-center text-emerald-400">
                <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <span className="text-emerald-400 text-sm font-semibold">To'lov yaratildi</span>
            </div>

            <div className="space-y-2.5">
              <div className="flex items-center justify-between text-sm">
                <span className="text-zinc-500">Summa</span>
                <span className="text-zinc-100 font-semibold font-mono">
                  {formatAmount(result.amount)} UZS
                </span>
              </div>

              {result.card && (
                <>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-zinc-500">Karta</span>
                    <span className="text-zinc-300 font-mono">*{result.card.last4}</span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-zinc-500">Egasi</span>
                    <span className="text-zinc-300">{result.card.owner_name}</span>
                  </div>
                </>
              )}

              {result.transaction_id && (
                <div className="flex items-center justify-between text-sm">
                  <span className="text-zinc-500">Transaction ID</span>
                  <span className="text-zinc-400 font-mono text-xs truncate max-w-[180px]">
                    {result.transaction_id}
                  </span>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
