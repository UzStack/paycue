# Paycue

Paycue — to'lovlarni avtomatlashtirish uchun open source **SaaS** dastur. To'lov
tizimlariga integratsiya qilmasdan, Humo kartangizga tushgan to'lovlarni Telegram
orqali aniqlab, sizning API'ngizga `webhook` yuboradi.

Bitta server bir nechta foydalanuvchiga (multi-tenant) xizmat qiladi. Har bir
foydalanuvchi API orqali ro'yxatdan o'tadi, o'z Telegram accountini ulaydi,
cartalarini qo'shadi va webhook manzilini sozlaydi.

> Savollar bo'lsa: [@Azamov_Samandar](https://t.me/Azamov_Samandar)

## Loyiha qanday ishlaydi?

1. Foydalanuvchi **ro'yxatdan o'tadi** → doimiy `token` oladi (tasdiq talab qilinmaydi).
2. **Telegram account** ulaydi (API orqali: telefon → SMS kod → kerak bo'lsa 2FA parol).
   Ulangach dastur avtomatik `@HUMOcardbot`ni topadi, `/start` bosilmagan bo'lsa bosadi va chatni kuzatadi.
3. **Carta** qo'shadi — cartaning oxirgi 4 raqami (`*7159`) bo'yicha.
4. **Webhook** URL sozlaydi.
5. To'lov kerak bo'lganda **transaction** yaratadi: `amount` + `card_id` yuboradi,
   dastur o'sha carta uchun hozir band bo'lmagan summani qaytaradi (masalan `20001`).
   Increment **har carta bo'yicha alohida** hisoblanadi.
6. Kartaga shu summada pul tushganda dastur webhook orqali xabar beradi.

## Komponentlar

- **`paycue`** — API server (Telegram watcher, worker, webhook).
- **`paycue-cli`** — API client (terminal). To'liq API orqali ishlaydi, dastur ichiga kirmaydi.

## Kerakli narsalar

1. Telegram account
2. Humo telegram bot ([@HUMOcardbot](https://t.me/HUMOcardbot))
3. Humo plastik kartasi
4. Server
5. Telegram `APP_ID` va `APP_HASH` ([my.telegram.org](https://my.telegram.org))

## Tez o'rnatish

```bash
curl -fsSL https://raw.githubusercontent.com/UzStack/paycue/main/install.sh | sudo bash
```

`install.sh`:
- o'rnatilmagan bo'lsa — `APP_ID`/`APP_HASH` so'raydi, `.env` tayyorlaydi, oxirgi
  releasedan binarylarni yuklaydi, systemd servisini sozlaydi, `paycue-cli`ni o'rnatadi;
- o'rnatilgan bo'lsa — binarylarni oxirgi releasega yangilab, servisni qayta ishga tushiradi.

## Qo'lda ishga tushirish

`.env` fayl (namuna `.env.example`da):

```bash
APP_ID=<app_id>
APP_HASH=<app_hash>
PORT=8080
DB_PATH=./db.sqlite3
SESSION_DIR=sessions
WORKERS=10
TRANSACTION_TIMEOUT=30
DEBUG=false
```

```bash
make build          # bin/paycue va bin/paycue-cli
./bin/paycue        # serverni ishga tushirish
```

## API

Barcha himoyalangan endpointlar `Authorization: Bearer <token>` headerini talab qiladi.

### Ro'yxatdan o'tish (public)

```bash
curl -X POST http://<host>:8080/api/register \
  -H 'content-type: application/json' \
  -d '{"name":"Ism Familiya","email":"pochta@example.com"}'
```

> `name` majburiy; `email` yoki `phone` dan kamida bittasi majburiy. Javobda doimiy `token` qaytadi.

### Webhook sozlash

```bash
curl -X POST http://<host>:8080/api/webhook \
  -H "Authorization: Bearer <token>" \
  -d '{"url":"https://example.com/hook"}'
```

Javobda `secret` qaytadi — dastur webhook yuborganda uni `X-API-Key` headerda
yuboradi. Callback URL'ingizda shu kalitni tekshirib, so'rov haqiqatan paycue'dan
kelganini bilib oling.

### Telegram account ulash

```bash
# 1) kod yuborish
curl -X POST http://<host>:8080/api/telegram/send-code \
  -H "Authorization: Bearer <token>" -d '{"phone":"+99890..."}'
# -> { "telegram_account_id": 1 }

# 2) kodni tasdiqlash (2FA bo'lsa password qo'shing)
curl -X POST http://<host>:8080/api/telegram/verify \
  -H "Authorization: Bearer <token>" \
  -d '{"telegram_account_id":1,"code":"12345"}'
# 2FA kerak bo'lsa javob: { "need_password": true } -> password bilan qayta yuboring
```

`GET /api/telegram` — accountlar ro'yxati.

### Carta qo'shish

```bash
curl -X POST http://<host>:8080/api/cards \
  -H "Authorization: Bearer <token>" \
  -d '{"telegram_account_id":1,"last4":"7159","label":"Asosiy"}'
```

`GET /api/cards` — cartalar ro'yxati.

### Transaction yaratish

```bash
curl -X POST http://<host>:8080/api/transactions \
  -H "Authorization: Bearer <token>" \
  -d '{"card_id":1,"amount":20000}'
```

```json
{ "status": true, "data": { "amount": 20001, "card_id": 1, "transaction_id": "<uuid>" } }
```

> `amount` — siz xohlagan summa; javobdagi `amount` — foydalanuvchidan so'raladigan
> (band bo'lmagan) summa. Increment har carta bo'yicha alohida hisoblanadi.

### Webhook payload

Kartaga to'lov tushganda yoki transaction bekor qilinganda dastur sizning URL'ingizga POST yuboradi:

```json
{ "action": "confirm", "amount": 20001, "card_id": 1, "transaction_id": "<uuid>" }
```

`action`: `confirm` (to'lov tushdi) yoki `cancel` (muddati o'tdi). Header'da
`X-API-Key: <secret>` bo'ladi. Callback URL `{ "ok": true }` va `200` qaytarishi kerak,
aks holda dastur `3 marta` qayta urinadi.

## CLI

```bash
paycue-cli register --name "Ism" --email pochta@example.com   # token saqlanadi
paycue-cli webhook --url https://example.com/hook
paycue-cli telegram send-code --phone +99890...
paycue-cli telegram verify --account 1 --code 12345 [--password 2FA]
paycue-cli telegram list
paycue-cli card add --account 1 --last4 7159 --label Asosiy
paycue-cli card list
paycue-cli transaction create --card 1 --amount 20000
```

Sozlash: `--api` (yoki `PAYCUE_API`), `--token` (yoki `PAYCUE_TOKEN`, yoki
`~/.config/paycue/token` — `register`dan keyin avtomatik saqlanadi).

## Muhim ma'lumotlar

- Transaction `TRANSACTION_TIMEOUT` daqiqa (default 30) active qoladi, keyin avtomatik bekor qilinadi.
- Har Telegram account uchun alohida session `SESSION_DIR`da saqlanadi.
- Ko'p transactionda summa farqi o'sib boradi; buni kamaytirish uchun bir nechta carta/account ishlating.

## Savollar?

- `Telegram accountni ulash xavfsizmi?` — Ha, dastur open source, ma'lumotlar o'z
  serveringizda qoladi, session fayllari sizda saqlanadi.
- `Yordam bera olamanmi?` — Albatta, fork qiling va pull request yuboring.
