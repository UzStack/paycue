import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api'
import { formatPhone, rawPhone } from '../format'

export default function Register() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ name: '', email: '', phone: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')

    if (!form.name.trim()) { setError('Ism majburiy'); return }
    if (!form.email.trim() && !rawPhone(form.phone)) {
      setError('Email yoki telefon raqamdan kamida biri kiritilishi shart')
      return
    }
    if (form.password && form.password.length < 6) {
      setError('Parol kamida 6 belgidan iborat bo\'lishi kerak')
      return
    }

    setLoading(true)
    try {
      const body = { name: form.name.trim() }
      if (form.email.trim()) body.email = form.email.trim()
      if (rawPhone(form.phone)) body.phone = rawPhone(form.phone)
      if (form.password) body.password = form.password

      const data = await api.register(body)
      localStorage.setItem('paycue_token', data.token)
      if (data.name || data.phone || data.email) {
        localStorage.setItem('paycue_user', JSON.stringify(data))
      }
      navigate('/dashboard')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <Link to="/" className="text-sky-400 font-bold text-xl tracking-tight">Paycue</Link>
          <h1 className="text-zinc-100 text-2xl font-bold mt-4 mb-1">Ro'yxatdan o'tish</h1>
          <p className="text-zinc-500 text-sm">Yangi hisob yarating</p>
        </div>

        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1.5">
                Ism <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="Ismingiz"
                className="w-full px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={loading}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1.5">
                Email <span className="text-zinc-500 font-normal">(ixtiyoriy)</span>
              </label>
              <input
                type="email"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                placeholder="email@example.com"
                className="w-full px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={loading}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1.5">
                Telefon <span className="text-zinc-500 font-normal">(ixtiyoriy)</span>
              </label>
              <input
                type="tel"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: formatPhone(e.target.value) })}
                placeholder="+998 90 123 45 67"
                className="w-full px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={loading}
              />
              <p className="text-xs text-zinc-500 mt-1">Email yoki telefon kamida biri talab qilinadi</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1.5">
                Parol <span className="text-zinc-500 font-normal">(ixtiyoriy, min. 6 belgi)</span>
              </label>
              <input
                type="password"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                placeholder="Parol kiriting"
                className="w-full px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={loading}
              />
            </div>

            {error && (
              <div className="px-3 py-2 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 text-sm">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 bg-sky-500 hover:bg-sky-400 disabled:bg-sky-500/50 disabled:cursor-not-allowed text-white font-medium rounded-md text-sm transition-colors"
            >
              {loading ? 'Yaratilmoqda...' : "Hisob yaratish"}
            </button>
          </form>

          <p className="text-center text-sm text-zinc-500 mt-4">
            Hisobingiz bormi?{' '}
            <Link to="/login" className="text-sky-400 hover:text-sky-300 transition-colors">
              Kirish
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
