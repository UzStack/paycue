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

`install.sh` (Linux — server + CLI):
- o'rnatilmagan bo'lsa — `APP_ID`/`APP_HASH` so'raydi, `.env` tayyorlaydi, oxirgi
  releasedan binarylarni yuklaydi, systemd servisini sozlaydi, `paycue-cli`ni o'rnatadi;
- o'rnatilgan bo'lsa — binarylarni oxirgi releasega yangilab, servisni qayta ishga tushiradi.

### Faqat CLI (macOS yoki masofaviy server uchun)

Faqat `paycue-cli`ni o'rnatish (serversiz). macOS'da avtomatik shu rejim tanlanadi
(server systemd talab qiladi, macOS'da bu yo'q):

```bash
curl -fsSL https://raw.githubusercontent.com/UzStack/paycue/main/install.sh | sudo bash -s -- --cli-only
```

> Server **Linux**da ishlaydi (SQLite/CGO). macOS uchun faqat CLI build qilinadi —
> u masofaviy serverga `--api` orqali ulanadi.

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

## API integratsiya

Bu — dasturchi boshqaradigan to'liq API. `paycue-cli` ham aynan shu
endpointlarni chaqiradi, ya'ni CLI'dagi har bir buyruq quyidagi so'rovga teng.

**Asoslar**
- Base URL: `http://<host>:<PORT>` (default port `8080`)
- Format: so'rov va javob — JSON (`Content-Type: application/json`)
- Auth: `POST /api/register` va `GET /health/` dan tashqari hammasi
  `Authorization: Bearer <token>` (yoki `X-API-Key: <token>`) talab qiladi
- Javob qobig'i: `{ "status": bool, "data": object }`. Xatoda
  `status=false` va `data.detail` matn bo'ladi (HTTP kodi: 400/401/403/404/5xx)

### Endpointlar jadvali

| Metod | Path | Auth | CLI ekvivalenti |
| --- | --- | --- | --- |
| `GET`  | `/health/` | ✗ | — |
| `POST` | `/api/register` | ✗ | `register` |
| `POST` | `/api/login` | ✗ | `login` |
| `GET`  | `/api/webhook` | ✓ | `webhook` (urlsiz) |
| `POST` | `/api/webhook` | ✓ | `webhook --url` |
| `POST` | `/api/telegram/send-code` | ✓ | `telegram connect` (1-qadam) |
| `POST` | `/api/telegram/verify` | ✓ | `telegram connect` (2-qadam) |
| `GET`  | `/api/telegram` | ✓ | `telegram list` |
| `POST` | `/api/cards` | ✓ | `card add` |
| `GET`  | `/api/cards` | ✓ | `card list` |
| `POST` | `/api/transactions` | ✓ | `transaction create` |

### So'rov/javob maydonlari

**`POST /api/register`** — ro'yxatdan o'tish, doimiy token oladi.

| Maydon | Tur | Majburiy | Izoh |
| --- | --- | --- | --- |
| `name` | string | ✓ | ism familiya |
| `email` | string | shartli | `email` yoki `phone` dan kamida bittasi |
| `phone` | string | shartli | — |
| `password` | string | ✓ | kamida 6 belgi (keyin login uchun) |

Javob: `{ "id": int, "name": string, "token": string }`

**`POST /api/login`** — email/phone + parol orqali tokenni qaytaradi.

| Maydon | Tur | Majburiy | Izoh |
| --- | --- | --- | --- |
| `login` | string | ✓ | email yoki phone |
| `password` | string | ✓ | — |

Javob: `{ "token": string }`. Noto'g'ri bo'lsa `401`.

**`POST /api/webhook`** — webhook URL sozlash.

| Maydon | Tur | Majburiy |
| --- | --- | --- |
| `url` | string | ✓ |

Javob: `{ "url": string, "secret": string }` — `secret` keyin webhookda `X-API-Key` sifatida keladi.

**`POST /api/telegram/send-code`** — Telegram account ulashni boshlaydi (SMS kod yuboradi).

| Maydon | Tur | Majburiy |
| --- | --- | --- |
| `phone` | string | ✓ |

Javob: `{ "telegram_account_id": int, "message": string }`

**`POST /api/telegram/verify`** — kodni (kerak bo'lsa 2FA parolni) tasdiqlaydi.

| Maydon | Tur | Majburiy |
| --- | --- | --- |
| `telegram_account_id` | int | ✓ |
| `code` | string | ✓ |
| `password` | string | 2FA yoqilgan bo'lsa |

Javob: 2FA kerak bo'lsa `{ "need_password": true }`; aks holda `{ "telegram_account_id": int, "status": "active" }`.

**`GET /api/telegram`** — accountlar ro'yxati. Javob: `TelegramAccount[]`.

**`POST /api/cards`** — carta qo'shish.

| Maydon | Tur | Majburiy | Izoh |
| --- | --- | --- | --- |
| `telegram_account_id` | int | ✓ | siznikilardan biri |
| `last4` | string | ✓ | aniq 4 raqam (`7159`) |
| `label` | string | ✗ | ixtiyoriy nom |

Javob: `Card` obyekti.

**`GET /api/cards`** — cartalar ro'yxati. Javob: `Card[]`.

**`POST /api/transactions`** — to'lov uchun band bo'lmagan summa oladi.

| Maydon | Tur | Majburiy | Izoh |
| --- | --- | --- | --- |
| `card_id` | int | ✓ | siznikilardan biri |
| `amount` | int | ✓ | so'ralayotgan summa (musbat) |

Javob: `{ "amount": int, "card_id": int, "transaction_id": string }` — `amount` band bo'lmagan summa (carta bo'yicha increment qilingan).

### Obyekt sxemalari

```jsonc
// TelegramAccount
{ "id": 1, "user_id": 1, "phone": "+99890...", "tg_user_id": 123, "username": "ali", "status": "active", "created_at": "..." }
// status: "pending" (kod kutilmoqda) | "active" (ulangan)

// Card
{ "id": 1, "telegram_account_id": 1, "last4": "7159", "label": "Asosiy", "created_at": "..." }
```

---

Quyida har bir endpoint uchun `curl` misollari.

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

API ikki qadamli (send-code → verify). CLI'da esa bu **bitta** `telegram connect`
buyrug'i — kodni va kerak bo'lsa 2FA parolni interaktiv so'raydi.

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

`paycue-cli` ni **argumentsiz** ishga tushirsangiz interaktiv menu (TUI) ochiladi —
menyularni **strelka (↑/↓) bilan** tanlab, Enter bilan tasdiqlaysiz (parol `*` bilan
yashiriladi):

```bash
paycue-cli            # interaktiv menu
```

Har bir amalni **ikki xil** ishlatish mumkin: interaktiv menu (TUI) yoki
to'g'ridan-to'g'ri subcommand (skript/avtomatlashtirish uchun qulay):

```bash
paycue-cli register --name "Ism" --email pochta@example.com --password "parol123"
paycue-cli login --login pochta@example.com --password "parol123"   # boshqa qurilmada token olish
paycue-cli webhook                              # joriy webhookni ko'rish
paycue-cli webhook --url https://example.com/hook   # webhook sozlash
paycue-cli telegram connect --phone +99890...   # interaktiv: kod va 2FA ni so'raydi
# yoki skriptbop (non-interaktiv) ikki qadam:
paycue-cli telegram send-code --phone +99890...               # -> telegram_account_id
paycue-cli telegram verify --account 1 --code 12345 [--password 2FA]
paycue-cli telegram list
paycue-cli card add --account 1 --last4 7159 --label Asosiy
paycue-cli card list
paycue-cli transaction create --card 1 --amount 20000
```

### Bir nechta account (profillar)

CLI bir nechta paycue accountni profillar orqali boshqaradi. Har profil o'z
`api` + `token`ini `~/.config/paycue/config.json`da saqlaydi.

```bash
paycue-cli register --name "Ali" --email ali@x.com --profile ali   # 'ali' profili
paycue-cli register --name "Vali" --phone +998... --profile vali   # 'vali' profili

paycue-cli profile list           # profillar (joriy * bilan belgilanadi)
paycue-cli profile current        # joriy profil
paycue-cli profile token [ali]    # profil tokenini chiqarish (default: joriy)
paycue-cli profile use ali        # joriy profilni almashtirish
paycue-cli profile add boss --token <token> [--api URL]   # tashqi token qo'shish
paycue-cli profile remove vali

# bitta buyruq uchun profilni almashtirmasdan tanlash:
paycue-cli --profile vali card list
```

Yechim tartibi (token/api uchun): `--token`/`--api` flag → tanlangan profil →
`PAYCUE_TOKEN`/`PAYCUE_API` env → default. Eski yagona `~/.config/paycue/token`
fayli ilk ishga tushishda avtomatik `default` profilga ko'chiriladi.

## Muhim ma'lumotlar

- Transaction `TRANSACTION_TIMEOUT` daqiqa (default 30) active qoladi, keyin avtomatik bekor qilinadi.
- Har Telegram account uchun alohida session `SESSION_DIR`da saqlanadi.
- Ko'p transactionda summa farqi o'sib boradi; buni kamaytirish uchun bir nechta carta/account ishlating.

## Savollar?

- `Telegram accountni ulash xavfsizmi?` — Ha, dastur open source, ma'lumotlar o'z
  serveringizda qoladi, session fayllari sizda saqlanadi.
- `Yordam bera olamanmi?` — Albatta, fork qiling va pull request yuboring.
