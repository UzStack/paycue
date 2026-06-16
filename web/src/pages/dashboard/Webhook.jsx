import { useEffect, useState } from 'react'
import { api } from '../../api'

export default function Webhook() {
  const [webhook, setWebhook] = useState(null)
  const [loading, setLoading] = useState(true)
  const [url, setUrl] = useState('')
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [copied, setCopied] = useState(false)
  const [secretVisible, setSecretVisible] = useState(false)

  function loadWebhook() {
    setLoading(true)
    api.getWebhook()
      .then((data) => {
        setWebhook(data)
        setUrl(data.url || '')
      })
      .catch(() => setWebhook(null))
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadWebhook() }, [])

  async function handleSave(e) {
    e.preventDefault()
    setSaveError('')
    setSaveSuccess(false)
    if (!url.trim()) { setSaveError('URL kiriting'); return }
    setSaving(true)
    try {
      const data = await api.setWebhook({ url: url.trim() })
      setWebhook(data)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch (err) {
      setSaveError(err.message)
    } finally {
      setSaving(false)
    }
  }

  async function handleCopy(text) {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setCopied(false)
    }
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-zinc-100 tracking-tight">Webhook</h1>
        <p className="text-zinc-400 text-sm mt-1">To'lov xabarnomalarini qabul qiling</p>
      </div>

      {/* Joriy webhook */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-5">
        <h2 className="text-sm font-semibold text-zinc-100 mb-4">Joriy sozlamalar</h2>

        {loading ? (
          <div className="space-y-3">
            <div className="h-8 bg-zinc-800 rounded animate-pulse" />
            <div className="h-8 bg-zinc-800 rounded animate-pulse w-3/4" />
          </div>
        ) : webhook && webhook.url ? (
          <div className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-zinc-500 mb-1.5 uppercase tracking-wider">
                Webhook URL
              </label>
              <div className="flex items-center gap-2 px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md">
                <span className="text-zinc-300 text-sm font-mono flex-1 break-all">{webhook.url}</span>
              </div>
            </div>

            {webhook.secret && (
              <div>
                <label className="block text-xs font-medium text-zinc-500 mb-1.5 uppercase tracking-wider">
                  Secret (X-API-Key)
                </label>
                <div className="flex items-center gap-2">
                  <div className="flex-1 flex items-center gap-2 px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md overflow-hidden">
                    <span className="text-zinc-300 text-sm font-mono flex-1 truncate">
                      {secretVisible ? webhook.secret : '••••••••••••••••••••••••'}
                    </span>
                    <button
                      onClick={() => setSecretVisible(!secretVisible)}
                      className="text-zinc-500 hover:text-zinc-300 transition-colors shrink-0"
                      title={secretVisible ? 'Yashirish' : 'Ko\'rish'}
                    >
                      {secretVisible ? (
                        <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                        </svg>
                      ) : (
                        <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                          <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                      )}
                    </button>
                  </div>
                  <button
                    onClick={() => handleCopy(webhook.secret)}
                    className="flex items-center gap-1.5 px-3 py-2.5 border border-zinc-700 hover:border-zinc-600 text-zinc-400 hover:text-zinc-200 rounded-md text-xs transition-colors shrink-0"
                  >
                    {copied ? (
                      <>
                        <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                        </svg>
                        Nusxalandi
                      </>
                    ) : (
                      <>
                        <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 01-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 011.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-8.876a9.06 9.06 0 00-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 01-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 00-3.375-3.375h-1.5a1.125 1.125 0 01-1.125-1.125v-1.5a3.375 3.375 0 00-3.375-3.375H9.75" />
                        </svg>
                        Nusxalash
                      </>
                    )}
                  </button>
                </div>
                <div className="mt-2 flex items-start gap-2 text-xs text-zinc-500 bg-zinc-800/50 px-3 py-2 rounded-md">
                  <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24" className="shrink-0 mt-0.5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
                  </svg>
                  Callback so'rovida <code className="text-zinc-400 bg-zinc-700 px-1 rounded mx-1">X-API-Key</code> header sifatida ushbu secretni tekshiring
                </div>
              </div>
            )}
          </div>
        ) : (
          <p className="text-zinc-500 text-sm">Webhook hali sozlanmagan</p>
        )}
      </div>

      {/* URL sozlash */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
        <h2 className="text-sm font-semibold text-zinc-100 mb-4">
          {webhook?.url ? 'URL yangilash' : 'Webhook sozlash'}
        </h2>

        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">
              Webhook URL
            </label>
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://your-server.com/webhook"
              className="w-full px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
              disabled={saving}
            />
            <p className="text-xs text-zinc-500 mt-1">
              To'lov yaratilganda ushbu URL'ga POST so'rov yuboriladi
            </p>
          </div>

          {saveError && (
            <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-md px-3 py-2">
              {saveError}
            </p>
          )}
          {saveSuccess && (
            <p className="text-sm text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 rounded-md px-3 py-2">
              Webhook muvaffaqiyatli saqlandi
            </p>
          )}

          <button
            type="submit"
            disabled={saving}
            className="px-5 py-2.5 bg-sky-500 hover:bg-sky-400 disabled:bg-sky-500/50 text-white font-medium rounded-md text-sm transition-colors"
          >
            {saving ? 'Saqlanmoqda...' : 'Saqlash'}
          </button>
        </form>
      </div>
    </div>
  )
}
