import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api'
import { formatLogin, rawLogin } from '../format'

export default function Login() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ login: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    if (!form.login.trim()) { setError('Email yoki telefon kiriting'); return }
    setLoading(true)
    try {
      const data = await api.login({ login: rawLogin(form.login), password: form.password })
      localStorage.setItem('paycue_token', data.token)
      navigate('/dashboard')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <Link to="/" className="text-sky-400 font-bold text-xl tracking-tight">Paycue</Link>
          <h1 className="text-zinc-100 text-2xl font-bold mt-4 mb-1">Kirish</h1>
          <p className="text-zinc-500 text-sm">Hisobingizga kiring</p>
        </div>

        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1.5">
                Email yoki telefon
              </label>
              <input
                type="text"
                value={form.login}
                onChange={(e) => setForm({ ...form, login: formatLogin(e.target.value) })}
                placeholder="+998 90 123 45 67 yoki email"
                className="w-full px-3 py-2.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500/30 transition-colors"
                disabled={loading}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1.5">
                Parol
              </label>
              <input
                type="password"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                placeholder="Parolingiz"
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
              {loading ? 'Kirilmoqda...' : 'Kirish'}
            </button>
          </form>

          <p className="text-center text-sm text-zinc-500 mt-4">
            Hisobingiz yo'qmi?{' '}
            <Link to="/register" className="text-sky-400 hover:text-sky-300 transition-colors">
              Ro'yxatdan o'ting
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
