import { Link } from 'react-router-dom'

function CodeBlock() {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-zinc-800">
        <div className="flex gap-1.5">
          <span className="w-3 h-3 rounded-full bg-zinc-700" />
          <span className="w-3 h-3 rounded-full bg-zinc-700" />
          <span className="w-3 h-3 rounded-full bg-zinc-700" />
        </div>
        <span className="text-xs text-zinc-500 font-mono ml-2">webhook payload</span>
      </div>
      <pre className="px-5 py-4 text-sm font-mono leading-relaxed text-zinc-300 overflow-x-auto">
        <span className="text-zinc-500">{'{'}</span>{'\n'}
        {'  '}<span className="text-sky-400">"amount"</span>
        <span className="text-zinc-500">: </span>
        <span className="text-emerald-400">50000</span>
        <span className="text-zinc-500">,</span>{'\n'}
        {'  '}<span className="text-sky-400">"card_last4"</span>
        <span className="text-zinc-500">: </span>
        <span className="text-amber-400">"4521"</span>
        <span className="text-zinc-500">,</span>{'\n'}
        {'  '}<span className="text-sky-400">"from"</span>
        <span className="text-zinc-500">: </span>
        <span className="text-amber-400">"Alisher T."</span>
        <span className="text-zinc-500">,</span>{'\n'}
        {'  '}<span className="text-sky-400">"timestamp"</span>
        <span className="text-zinc-500">: </span>
        <span className="text-amber-400">"2026-06-17T14:32:01Z"</span>{'\n'}
        <span className="text-zinc-500">{'}'}</span>
      </pre>
    </div>
  )
}

const STEPS = [
  {
    num: '01',
    title: "Ro'yxatdan o'tish",
    desc: "Ism va telefon raqam bilan tezda hisob yarating.",
  },
  {
    num: '02',
    title: 'Telegram ulash',
    desc: 'Telefon raqamingizga SMS kod keladi, uni kiritib accountni aktivlashtiring.',
  },
  {
    num: '03',
    title: 'Humo karta qo\'shish',
    desc: 'Karta raqami va egasining ismini kiriting. Bir nechta karta qo\'shishingiz mumkin.',
  },
  {
    num: '04',
    title: 'Webhook va tranzaksiya',
    desc: "To'lov URL sozlang. Karta orqali to'lov yaratilganda sizning serveringizga xabar keladi.",
  },
]

const USE_CASES = [
  {
    title: 'Onlayn do\'konlar',
    desc: "Telegram bot yoki saytdagi buyurtmalar uchun karta to'lovini avtomatik tasdiqlang — mijoz to'lov qilishi bilanoq buyurtma o'tadi.",
  },
  {
    title: 'Telegram botlar',
    desc: "Obuna, raqamli mahsulot yoki xizmat sotuvchi botlarga to'lov qabul qilishni ulang. Webhook kelishi bilan kontentni avtomatik bering.",
  },
  {
    title: 'Freelancer va xizmatlar',
    desc: "Mijozdan oldindan to'lov qabul qiling — har bir to'lov o'ziga xos summa bilan ajratiladi, kim to'laganini aniq bilasiz.",
  },
  {
    title: 'SaaS va obunalar',
    desc: "Oylik to'lovlarni kuzating. To'lov tushganda foydalanuvchi tarifini API orqali avtomatik faollashtiring.",
  },
]

const FAQ = [
  {
    q: 'Paycue qanday ishlaydi?',
    a: "Telegram accountingiz @HUMOcardbot orqali kartaga tushgan to'lovlarni real vaqtda kuzatadi. To'lov aniqlanganda Paycue sizning webhook URL'ingizga POST so'rov yuboradi — tizimingiz buyurtmani avtomatik tasdiqlaydi.",
  },
  {
    q: 'To\'lov tizimiga (Payme/Click) integratsiya kerakmi?',
    a: "Yo'q. Paycue rasmiy to'lov shlyuzini, shartnoma yoki merchant akkauntni talab qilmaydi. Oddiy Humo plastik kartasi va Telegram accounti yetarli.",
  },
  {
    q: 'Telegram accountni ulash xavfsizmi?',
    a: "Ha. Paycue to'liq open source — kodni o'zingiz ko'rishingiz mumkin. Barcha ma'lumotlar o'z serveringizda qoladi, Telegram session fayllari ham sizda saqlanadi, uchinchi tomonga hech narsa yuborilmaydi.",
  },
  {
    q: 'Bir vaqtda ikki kishi bir xil summa to\'lasa nima bo\'ladi?',
    a: "Paycue har bir transaction uchun band bo'lmagan noyob summani (masalan 20001, 20002) ajratadi. Shu sababli to'lovlar bir-biriga aralashmaydi va to'g'ri tranzaksiyaga bog'lanadi.",
  },
  {
    q: 'Bir nechta karta ishlatsam bo\'ladimi?',
    a: "Ha. Bir nechta Telegram account va Humo karta qo'shishingiz mumkin. Karta ko'rsatmasangiz, Paycue eng kam yuklangan kartani avtomatik tanlab, yukni teng taqsimlaydi.",
  },
  {
    q: 'Paycue bepulmi?',
    a: "Ha, Paycue open source. O'z serveringizda bepul ishga tushirasiz. Bitta server bir nechta foydalanuvchiga (multi-tenant) xizmat qiladi.",
  },
]

const FEATURES = [
  {
    title: 'Real vaqtda monitoring',
    desc: 'Telegram akkauntingiz orqali Humo kartaga tushgan har bir to\'lovni darhol aniqlaydi.',
    icon: (
      <svg width="22" height="22" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
      </svg>
    ),
  },
  {
    title: 'Avtomatik karta tanlash',
    desc: "Karta ko'rsatmasangiz, eng kam yukli karta avtomatik tanlanadi - yukni teng taqsimlaydi.",
    icon: (
      <svg width="22" height="22" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" />
      </svg>
    ),
  },
  {
    title: 'Xavfsiz webhook',
    desc: 'Har bir webhook so\'rovi X-API-Key header orqali imzolanadi. Soxta so\'rovlar bloklanadi.',
    icon: (
      <svg width="22" height="22" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
      </svg>
    ),
  },
  {
    title: 'Ko\'p karta va akkaunt',
    desc: 'Bir nechta Telegram akkaunt va Humo karta bilan ishlang. API orqali to\'liq boshqarish.',
    icon: (
      <svg width="22" height="22" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5M12 17.25h8.25" />
      </svg>
    ),
  },
]

export default function Landing() {
  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100">
      {/* Nav */}
      <nav className="sticky top-0 z-50 border-b border-zinc-800/60 bg-zinc-950/90 backdrop-blur-sm">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 flex items-center justify-between h-16">
          <span className="text-sky-400 font-bold text-lg tracking-tight">Paycue</span>
          <div className="flex items-center gap-3">
            <Link
              to="/login"
              className="text-sm text-zinc-400 hover:text-zinc-100 px-4 py-2 rounded-md border border-zinc-700 hover:border-zinc-500 transition-colors"
            >
              Kirish
            </Link>
            <Link
              to="/register"
              className="text-sm text-zinc-100 bg-sky-500 hover:bg-sky-400 px-4 py-2 rounded-md font-medium transition-colors"
            >
              Boshlash
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero - asimmetrik split */}
      <section className="max-w-6xl mx-auto px-4 sm:px-6 py-20 lg:py-28">
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-16 items-center">
          {/* Chap: matn */}
          <div>
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 text-xs font-medium mb-6">
              <span className="w-1.5 h-1.5 rounded-full bg-sky-400" />
              Telegram + Humo integratsiya
            </div>
            <h1 className="text-4xl sm:text-5xl lg:text-5xl font-bold tracking-tight leading-tight text-zinc-50 mb-5">
              Telegram orqali{' '}
              <span className="text-sky-400">to'lovlarni</span>{' '}
              avtomatlashtir
            </h1>
            <p className="text-zinc-400 text-lg leading-relaxed mb-8 max-w-md">
              Humo kartaga tushgan to'lovlarni aniqlang va webhook orqali tizimingizga yetkazing.
            </p>
            <div className="flex flex-wrap gap-3">
              <Link
                to="/register"
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-sky-500 hover:bg-sky-400 text-white font-medium rounded-md text-sm transition-colors"
              >
                Bepul boshlash
                <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
                </svg>
              </Link>
              <a
                href="https://github.com/UzStack/paycue"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-5 py-2.5 border border-zinc-700 hover:border-zinc-500 text-zinc-300 hover:text-zinc-100 font-medium rounded-md text-sm transition-colors"
              >
                <svg width="16" height="16" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
                </svg>
                GitHub
              </a>
            </div>
          </div>

          {/* O'ng: kod bloki */}
          <div className="relative">
            <div className="absolute -inset-4 bg-sky-500/5 rounded-xl blur-xl" />
            <div className="relative">
              <CodeBlock />
              <div className="mt-3 flex items-center gap-2 text-xs text-zinc-500 px-1">
                <svg width="12" height="12" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
                </svg>
                To'lov yaratilganda sizning serveringizga yuboriladi
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Qanday ishlaydi - zig-zag steps */}
      <section className="border-t border-zinc-800/60 bg-zinc-900/30">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-20">
          <div className="mb-12">
            <h2 className="text-2xl sm:text-3xl font-bold text-zinc-50 tracking-tight mb-3">
              Qanday ishlaydi
            </h2>
            <p className="text-zinc-400 text-base max-w-lg">
              To'rt qadamda integratsiyani yoqing va to'lovlarni avtomatik kuzating.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-px bg-zinc-800/60 rounded-lg overflow-hidden">
            {STEPS.map((step) => (
              <div key={step.num} className="bg-zinc-900 p-6 hover:bg-zinc-800/70 transition-colors">
                <span className="text-sky-500 font-mono text-xs font-bold mb-4 block">{step.num}</span>
                <h3 className="text-zinc-100 font-semibold text-base mb-2">{step.title}</h3>
                <p className="text-zinc-400 text-sm leading-relaxed">{step.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Xususiyatlar - bento 2x2 */}
      <section className="max-w-6xl mx-auto px-4 sm:px-6 py-20">
        <div className="mb-12">
          <h2 className="text-2xl sm:text-3xl font-bold text-zinc-50 tracking-tight mb-3">
            Nima beradi
          </h2>
          <p className="text-zinc-400 text-base max-w-lg">
            Ishlab chiqish uchun to'g'ridan-to'g'ri API, hech qanday murakkab integratsiyasiz.
          </p>
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          {FEATURES.map((f) => (
            <div
              key={f.title}
              className="group bg-zinc-900 border border-zinc-800 rounded-lg p-6 hover:border-zinc-700 transition-colors"
            >
              <div className="w-10 h-10 rounded-md bg-sky-500/10 border border-sky-500/20 flex items-center justify-center text-sky-400 mb-4 group-hover:bg-sky-500/15 transition-colors">
                {f.icon}
              </div>
              <h3 className="text-zinc-100 font-semibold text-base mb-2">{f.title}</h3>
              <p className="text-zinc-400 text-sm leading-relaxed">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Kimlar uchun - use cases */}
      <section className="border-t border-zinc-800/60 bg-zinc-900/30">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-20">
          <div className="mb-12">
            <h2 className="text-2xl sm:text-3xl font-bold text-zinc-50 tracking-tight mb-3">
              Kimlar uchun
            </h2>
            <p className="text-zinc-400 text-base max-w-lg">
              Karta orqali to'lov qabul qiladigan va uni avtomatik tasdiqlashni xohlaydigan har qanday loyiha uchun.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {USE_CASES.map((u) => (
              <div key={u.title} className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 hover:border-zinc-700 transition-colors">
                <h3 className="text-zinc-100 font-semibold text-base mb-2">{u.title}</h3>
                <p className="text-zinc-400 text-sm leading-relaxed">{u.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* FAQ - tez-tez beriladigan savollar */}
      <section className="max-w-6xl mx-auto px-4 sm:px-6 py-20">
        <div className="mb-12">
          <h2 className="text-2xl sm:text-3xl font-bold text-zinc-50 tracking-tight mb-3">
            Tez-tez beriladigan savollar
          </h2>
          <p className="text-zinc-400 text-base max-w-lg">
            Paycue haqida eng ko'p so'raladigan savollar va javoblar.
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-x-8 gap-y-2">
          {FAQ.map((item) => (
            <details
              key={item.q}
              className="group border-b border-zinc-800 py-4"
            >
              <summary className="flex items-center justify-between cursor-pointer list-none text-zinc-100 font-medium text-base">
                {item.q}
                <svg
                  className="w-5 h-5 text-zinc-500 shrink-0 ml-4 transition-transform group-open:rotate-180"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                </svg>
              </summary>
              <p className="text-zinc-400 text-sm leading-relaxed mt-3 pr-8">{item.a}</p>
            </details>
          ))}
        </div>
      </section>

      {/* CTA banner */}
      <section className="border-t border-zinc-800/60">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-16">
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl px-8 py-10 sm:flex sm:items-center sm:justify-between gap-8">
            <div>
              <h2 className="text-xl sm:text-2xl font-bold text-zinc-50 tracking-tight mb-2">
                Bugun boshlang
              </h2>
              <p className="text-zinc-400 text-sm max-w-sm">
                Ro'yxatdan o'ting va birinchi webhook so'rovingizni daqiqalar ichida oling.
              </p>
            </div>
            <div className="flex flex-col sm:flex-row gap-3 mt-6 sm:mt-0 shrink-0">
              <Link
                to="/register"
                className="px-5 py-2.5 bg-sky-500 hover:bg-sky-400 text-white font-medium rounded-md text-sm transition-colors text-center"
              >
                Bepul boshlash
              </Link>
              <Link
                to="/login"
                className="px-5 py-2.5 border border-zinc-700 hover:border-zinc-500 text-zinc-300 hover:text-zinc-100 font-medium rounded-md text-sm transition-colors text-center"
              >
                Kirish
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-zinc-800/60">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-8 flex flex-col sm:flex-row items-center justify-between gap-4">
          <span className="text-sky-400 font-bold tracking-tight">Paycue</span>
          <p className="text-zinc-500 text-xs">
            &copy; 2026 Paycue. Telegram + Humo to'lov integratsiyasi.
          </p>
          <div className="flex gap-4 text-xs text-zinc-500">
            <Link to="/login" className="hover:text-zinc-300 transition-colors">Kirish</Link>
            <Link to="/register" className="hover:text-zinc-300 transition-colors">Ro'yxat</Link>
          </div>
        </div>
      </footer>
    </div>
  )
}
