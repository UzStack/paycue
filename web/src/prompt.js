// AI dasturchiga beriladigan Paycue integratsiya prompti.
// Foydalanuvchiga ko'rsatilmaydi — faqat clipboardga nusxalanadi.

export const REPO = 'https://github.com/UzStack/paycue'

export function buildIntegrationPrompt({ base, token, secret, webhookUrl }) {
  const secretLine = secret
    ? `- Webhook secret (X-API-Key, maxfiy): ${secret}`
    : `- Webhook secret: hali sozlanmagan — avval Paycue dashboardidan webhook URL kiriting, so'ng secret paydo bo'ladi.`
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
