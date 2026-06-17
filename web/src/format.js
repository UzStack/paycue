// Input maskalari/formatterlari — barcha formalarda bir xil ishlatiladi.
// Har biri: format* (ko'rsatish uchun) va raw* (serverga yuborish uchun) juftligi.

// Telefon — O'zbekiston: +998 XX XXX XX XX
export function formatPhone(value) {
  let d = String(value).replace(/\D/g, '')
  if (d.startsWith('998')) d = d.slice(3)
  d = d.slice(0, 9) // operator (2) + 7 raqam
  if (!d) return ''
  const parts = ['+998', d.slice(0, 2)]
  if (d.length > 2) parts.push(d.slice(2, 5))
  if (d.length > 5) parts.push(d.slice(5, 7))
  if (d.length > 7) parts.push(d.slice(7, 9))
  return parts.join(' ')
}

export function rawPhone(value) {
  const d = String(value).replace(/\D/g, '')
  if (!d) return ''
  const local = d.startsWith('998') ? d.slice(3) : d
  return '+998' + local.slice(0, 9)
}

// Karta raqami — 16 raqam, 4 talab guruh: 9860 1234 5678 9012
export function formatCardNumber(value) {
  const d = String(value).replace(/\D/g, '').slice(0, 16)
  return d.replace(/(.{4})/g, '$1 ').trim()
}

export function rawCard(value) {
  return String(value).replace(/\D/g, '').slice(0, 16)
}

// Summa — minglik ajratgich bilan: 50 000
export function formatAmount(value) {
  const d = String(value).replace(/\D/g, '')
  if (!d) return ''
  return Number(d).toLocaleString('uz-UZ')
}

export function rawAmount(value) {
  return String(value).replace(/\D/g, '')
}

// Tasdiqlash kodi — faqat raqam, maksimal 6 ta
export function formatCode(value) {
  return String(value).replace(/\D/g, '').slice(0, 6)
}

// Login maydoni — email yoki telefon. Faqat raqam/+/probel bo'lsa telefon deb
// formatlanadi, aks holda (email) o'zgartirilmaydi.
const phoneLike = /^[+\d][\d\s+]*$/

export function formatLogin(value) {
  const v = String(value)
  return v.trim() && phoneLike.test(v.trim()) ? formatPhone(v) : v
}

export function rawLogin(value) {
  const v = String(value).trim()
  return v && phoneLike.test(v) ? rawPhone(v) : v
}
