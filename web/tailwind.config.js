/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,jsx}"],
  theme: {
    extend: {
      fontFamily: { sans: ['Geist', 'system-ui', 'sans-serif'] },
      colors: {
        accent: { DEFAULT: '#0ea5e9', dark: '#0284c7', light: '#38bdf8' }
      }
    }
  },
  plugins: []
}
