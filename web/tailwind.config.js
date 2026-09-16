export default {
  presets: [require('cheval-ui/tailwind-preset')],
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{ts,tsx,js,jsx}',
    './node_modules/cheval-ui/dist/**/*.js',
  ],
  plugins: [require('tailwindcss-animate')],
}
