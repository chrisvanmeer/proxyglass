/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/server/**/*.go"],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
