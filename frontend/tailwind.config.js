/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        // JITRACK Brand Colors - Orange Theme
        'jit': {
          // Primary oranges
          orange: {
            50: '#fff7ed',
            100: '#ffedd5',
            200: '#fed7aa',
            300: '#fdba74',
            400: '#fb923c',
            500: '#f97316',
            600: '#ea580c',
            700: '#c2410c',
            800: '#9a3412',
            900: '#7c2d12',
          },
          // Dark mode backgrounds
          dark: {
            primary: '#0f0a08',      // Very dark brown-black
            secondary: '#1c1410',    // Dark brown
            card: '#18100c',         // Darker brown for cards
            surface: '#231915',      // Slightly lighter for surfaces
          },
          // Light mode backgrounds
          light: {
            primary: '#faf6f4',      // Warm off-white
            secondary: '#ffffff',    // Pure white
            card: '#fff8f5',         // Very light orange tint
            surface: '#fff0eb',      // Light peach
          },
          // Accent colors
          accent: '#f97316',         // Main orange
          accentHover: '#ea580c',    // Darker orange for hover
          success: '#22c55e',        // Green for success
          warning: '#eab308',        // Yellow for warnings
          error: '#ef4444',          // Red for errors
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      boxShadow: {
        'glow-orange': '0 0 20px rgba(249, 115, 22, 0.3)',
        'glow-orange-lg': '0 0 40px rgba(249, 115, 22, 0.4)',
      }
    },
  },
  darkMode: 'class',
  plugins: [],
}
