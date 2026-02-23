/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Apple iOS/macOS 色板
        primary: {
          50: '#eef4ff',
          100: '#d9e7ff',
          200: '#bbcfff',
          300: '#8db0ff',
          400: '#5888ff',
          500: '#007AFF', // iOS Blue
          600: '#0056CC',
          700: '#0040A0',
          800: '#003580',
          900: '#002966',
        },
        // Apple 系统灰
        'apple-gray': {
          50: '#f5f5f7',
          100: '#e8e8ed',
          200: '#d2d2d7',
          300: '#b0b0b6',
          400: '#86868b',
          500: '#6e6e73',
          600: '#424245',
          700: '#333336',
          800: '#1d1d1f',
          900: '#0a0a0a',
        },
        // Apple 系统色
        'apple-green': '#34C759',
        'apple-red': '#FF3B30',
        'apple-orange': '#FF9500',
        'apple-yellow': '#FFCC00',
        'apple-teal': '#5AC8FA',
        'apple-purple': '#AF52DE',
        'apple-pink': '#FF2D55',
        'apple-indigo': '#5856D6',
      },
      borderRadius: {
        'apple': '12px',
        'apple-lg': '16px',
        'apple-xl': '20px',
      },
      boxShadow: {
        'apple-sm': '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)',
        'apple': '0 2px 8px rgba(0, 0, 0, 0.08), 0 1px 3px rgba(0, 0, 0, 0.04)',
        'apple-md': '0 4px 16px rgba(0, 0, 0, 0.08), 0 2px 6px rgba(0, 0, 0, 0.04)',
        'apple-lg': '0 8px 32px rgba(0, 0, 0, 0.1), 0 4px 12px rgba(0, 0, 0, 0.06)',
        'apple-xl': '0 16px 48px rgba(0, 0, 0, 0.12), 0 8px 24px rgba(0, 0, 0, 0.08)',
      },
      fontFamily: {
        'apple': ['-apple-system', 'BlinkMacSystemFont', 'SF Pro Display', 'SF Pro Text', 'Helvetica Neue', 'PingFang SC', 'Microsoft YaHei', 'sans-serif'],
      },
      backdropBlur: {
        'apple': '20px',
        'apple-lg': '40px',
      },
    },
  },
  plugins: [],
  corePlugins: {
    preflight: false, // 禁用Tailwind的预设样式，避免与Naive UI冲突
  },
}
