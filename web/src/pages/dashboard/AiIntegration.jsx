import { useEffect, useMemo, useState } from 'react'
import { api, getToken } from '../../api'

const REPO = 'https://github.com/UzStack/paycue'

function buildPrompt({ base, token, secret, webhookUrl }) {
  const secretLine = secret
    ? `- Webhook secret (X-API-Key, maxfiy): ${secret}`
    : `- Webhook secret: hali sozlanmagan — avval Paycue dashboardidan webhook URL kiriting, so'ng secret shu yerda paydo bo'ladi.`
  const webhookLine = webhookUrl
    ? `- Mening webhook URL'im (Paycue shu yerga POST yuboradi): ${webhookUrl}`
    : `- Webhook URL: hali sozlanmagan.`

  return `Sen mening dasturimga Paycue to'lov tizimini integratsiya qiladigan tajribali backend dasturchisan.

# Paycue nima
Paycue — Humo kartaga tushgan to'lovlarni Telegram orqali real vaqtda aniqlab, webhook
orqali mening serverimga xabar beradigan to'lov tizimi. Rasmiy to'lov shlyuzi (Payme/Click)
yoki merchant shartnoma talab qilmaydi. To'liq API hujjati va manba kodi: ${REPO}

# Konfiguratsiya
- API base URL: ${base}
- API token (maxfiy): ${token || '<dashboarddan token oling>'}
- Har bir himoyalangan so'rovda header: Authorization: Bearer <token>
- Javob qobig'i: { "status": bool, "data": object }. Xatoda status=false va data.detail matn bo'ladi.
${secretLine}
${webhookLine}

# Men xohlagan to'lov oqimi
1. Mijoz buyurtma bersa, backendim Paycue'da transaction yaratadi.
2. Paycue band bo'lmagan NOYOB summa va pay_url qaytaradi.
3. Mijozga pay_url havolasini ko'rsataman (yoki o'z UI'mda summa + kartani ko'rsataman).
4. Mijoz AYNAN shu summani ko'rsatilgan Humo kartaga o'tkazadi.
5. Pul tushganda Paycue mening webhook URL'imga POST yuboradi → men buyurtmani tasdiqlayman.

# Kerakli endpointlar

## 1) Transaction yaratish
POST ${base}/api/transactions
Header: Authorization: Bearer <token>
Body: { "amount": <so'ralayotgan summa, butun son>, "card_id": <ixtiyoriy> }
Javob (data): { "amount": <to'lanadigan NOYOB summa>, "card_id", "transaction_id", "pay_url" }
MUHIM: mijozdan aynan javobdagi "amount" ni so'ra — summa noyob, to'lov aynan shu orqali aniqlanadi.

## 2) Webhook (men qabul qilaman)
Paycue to'lov tushganda yoki muddati o'tganda mening URL'imga POST yuboradi:
  body: { "action": "confirm" | "cancel", "amount": int, "card_id": int, "transaction_id": string }
  header: X-API-Key: <secret>
Mening webhook handlerim shularni qilsin:
- Kelgan X-API-Key ni yuqoridagi secret bilan solishtirib tekshirsin (mos kelmasa 403 qaytarib rad etsin — soxta so'rovlardan himoya).
- transaction_id orqali tegishli buyurtmani topsin.
- action="confirm" bo'lsa buyurtmani "to'langan" deb belgilasin; action="cancel" bo'lsa "bekor qilingan".
- Javobda 200 status va {"ok": true} qaytarsin. Aks holda Paycue 3 martagacha qayta uradi.

## 3) Pay sahifa ma'lumoti (ixtiyoriy — o'z UI'ng uchun)
GET ${base}/api/pay/{transaction_id}   (auth shart emas)
Javob (data): { amount, card_number, card_owner, state, expires_at, timeout_mins }
state: active | confirmed | cancelled | expired. Bu bilan o'z to'lov sahifangda summa, karta va qolgan vaqtni ko'rsatishing mumkin.

# Vazifa
Mening dasturimga (texnologiya stekini mendan so'ra yoki mavjud kodimni tahlil qil) quyidagilarni qo'sh:
1. Buyurtma yaratilganda Paycue'da transaction yaratadigan funksiya/servis.
2. Mijozga to'lov summasi (javobdagi amount) va pay_url ni ko'rsatadigan oqim.
3. Webhook qabul qiluvchi endpoint — X-API-Key tekshiruvi va {"ok":true} javobi bilan.
4. Webhook kelganda buyurtma statusini yangilab, biznes-logikani (masalan mahsulotni berish) ishga tushir.
Token va secretni .env / sozlama faylida saqla va koddan o'qi — hech qachon manba kodga yozib qo'yma.
Avval qisqa integratsiya rejasini ko'rsat, so'ng kodni yoz.`
}

function PromptView({ value }) {
  const [copied, setCopied] = useState(false)
  function copy() {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1800)
    })
  }
  return (
    <div className="bg-zinc-950 border border-zinc-800 rounded-lg overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2.5 border-b border-zinc-800 bg-zinc-900/60">
        <span className="text-xs text-zinc-500 font-mono">integration-prompt.txt</span>
        <button
          onClick={copy}
          className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
            copied ? 'bg-emerald-500/15 text-emerald-400' : 'bg-sky-500 hover:bg-sky-400 text-white'
          }`}
        >
          {copied ? (
            <>
              <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
              Nusxalandi
            </>
          ) : (
            <>
              <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75c0-.212.03-.418.084-.612m7.332 0A48.2 48.2 0 0117.66 4.07c1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185" /></svg>
              Butun promptni nusxalash
            </>
          )}
        </button>
      </div>
      <pre className="px-4 py-4 text-xs leading-relaxed text-zinc-300 font-mono whitespace-pre-wrap break-words max-h-[480px] overflow-y-auto">
        {value}
      </pre>
    </div>
  )
}

export default function AiIntegration() {
  const token = getToken() || ''
  const [webhook, setWebhook] = useState({ url: '', secret: '' })
  const [loading, setLoading] = useState(true)

  const base = import.meta.env.VITE_API_URL || (typeof window !== 'undefined' ? window.location.origin : '')

  useEffect(() => {
    api.getWebhook()
      .then((d) => setWebhook({ url: d?.url || '', secret: d?.secret || '' }))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const prompt = useMemo(
    () => buildPrompt({ base, token, secret: webhook.secret, webhookUrl: webhook.url }),
    [base, token, webhook.secret, webhook.url]
  )

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-zinc-100 tracking-tight">AI integratsiya</h1>
        <p className="text-zinc-400 text-sm mt-1">
          Tayyor promptni AI dasturchiga (Claude, ChatGPT, Cursor) bering — u Paycue to'lovni dasturingizga integratsiya qiladi.
        </p>
      </div>

      {/* Qadamlar */}
      <div className="grid sm:grid-cols-3 gap-3 mb-6">
        {[
          { n: '1', t: 'Promptni nusxalang', d: 'Quyidagi prompt token va sozlamalaringiz bilan tayyor.' },
          { n: '2', t: 'AI ga joylang', d: 'Claude / ChatGPT / Cursor ga qo\'ying va dasturingizni ulang.' },
          { n: '3', t: 'Integratsiya', d: 'AI transaction yaratish va webhook qabul qilishni yozadi.' },
        ].map((s) => (
          <div key={s.n} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <span className="inline-flex items-center justify-center w-6 h-6 rounded-md bg-sky-500/10 text-sky-400 text-xs font-bold mb-2">{s.n}</span>
            <p className="text-zinc-100 text-sm font-medium">{s.t}</p>
            <p className="text-zinc-500 text-xs mt-1 leading-relaxed">{s.d}</p>
          </div>
        ))}
      </div>

      {/* Maxfiylik ogohlantirishi */}
      <div className="flex items-start gap-2 text-xs text-amber-400/90 bg-amber-500/5 border border-amber-500/15 rounded-md px-3 py-2.5 mb-4">
        <svg width="15" height="15" className="mt-px shrink-0" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" /></svg>
        Bu prompt sizning maxfiy <b>token</b>ingiz{webhook.secret ? ' va webhook secret' : ''}ini o'z ichiga oladi. Faqat ishonchli AI vositalariga bering, ommaga ulashmang.
      </div>

      {/* Webhook sozlanmagan ogohlantirishi */}
      {!loading && !webhook.url && (
        <div className="flex items-start gap-2 text-xs text-zinc-400 bg-zinc-900 border border-zinc-800 rounded-md px-3 py-2.5 mb-4">
          <svg width="15" height="15" className="mt-px shrink-0 text-sky-400" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" /></svg>
          To'liq integratsiya (webhook secret bilan) uchun avval <b className="mx-1">Webhook</b> sahifasidan callback URL'ingizni sozlang.
        </div>
      )}

      {/* Prompt */}
      {loading ? (
        <div className="h-80 bg-zinc-900 border border-zinc-800 rounded-lg animate-pulse" />
      ) : (
        <PromptView value={prompt} />
      )}

      {/* Repo havolasi */}
      <div className="mt-4 flex flex-wrap items-center gap-3">
        <a
          href={REPO}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 px-4 py-2 border border-zinc-700 hover:border-zinc-500 text-zinc-300 hover:text-zinc-100 rounded-md text-sm transition-colors"
        >
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
          </svg>
          GitHub repo — to'liq API hujjati
        </a>
        <span className="text-zinc-600 text-xs font-mono">{REPO}</span>
      </div>
    </div>
  )
}
