# Paycue Web

Paycue Go backend uchun React frontend (Vite + Tailwind).

## Ishga tushirish

```bash
cd web

# .env fayl yarating
cp .env.example .env
# VITE_API_URL ni kerak bo'lsa o'zgartiring

# Paketlarni o'rnating
npm install

# Dev server
npm run dev
# http://localhost:5173

# Production build
npm run build
# dist/ papkasida tayyor fayllar

# Build preview
npm run preview
```

## Sahifalar

| Yo'l | Tavsif |
|------|--------|
| `/` | Landing page |
| `/login` | Kirish |
| `/register` | Ro'yxatdan o'tish |
| `/dashboard` | Bosh sahifa (himoyalangan) |
| `/dashboard/telegram` | Telegram akkauntlar |
| `/dashboard/cards` | Kartalar |
| `/dashboard/webhook` | Webhook sozlamalar |
| `/dashboard/transaction` | Tranzaksiya yaratish |

## Muhit o'zgaruvchilari

| O'zgaruvchi | Default | Tavsif |
|-------------|---------|--------|
| `VITE_API_URL` | `http://127.0.0.1:8080` | Backend API URL |
