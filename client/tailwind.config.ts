import type { Config } from 'tailwindcss';
import defaultTheme from 'tailwindcss/defaultTheme';

export default {
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        primary: ['Inter', ...defaultTheme.fontFamily.sans],
      },
      colors: {
        'rwc-violet': {
          50: '#EAE6EF',
          100: '#D5CDDF',
          200: '#CBC1D7',
          300: '#B6A8C7',
          400: '#A18FB7',
          500: '#8C76A7',
          600: '#785F95',
          700: '#65507C',
          800: '#514064',
          900: '#3C304B',
          950: '#282032',
        },
        'rwc-cyan': {
          50: '#D5ECEC',
          100: '#B9DFDF',
          200: '#9DD2D2',
          300: '#81C5C5',
          400: '#65B8B8',
          500: '#4DA8A8',
          600: '#408C8C',
          700: '#374151',
          800: '#275454',
          900: '#1A3838',
          950: '#132A2A',
        },
        'rwc-beige': {
          50: '#F4F4E1',
          100: '#E9E9C3',
          200: '#DEDEA6',
          300: '#D3D388',
          400: '#C8C86A',
          500: '#BDBD4C',
          600: '#A4A43D',
          700: '#868632',
          800: '#686827',
          900: '#4A4A1C',
          950: '#3B3B16',
        },
        dark: '#222222',
        'custom-yellow': '#BA333',
      },
      keyframes: {
        flicker: {
          '0%, 19.999%, 22%, 62.999%, 64%, 64.999%, 70%, 100%': {
            opacity: '0.99',
            filter:
              'drop-shadow(0 0 1px rgba(252, 211, 77)) drop-shadow(0 0 15px rgba(245, 158, 11)) drop-shadow(0 0 1px rgba(252, 211, 77))',
          },
          '20%, 21.999%, 63%, 63.999%, 65%, 69.999%': {
            opacity: '0.4',
            filter: 'none',
          },
        },
        shimmer: {
          '0%': {
            backgroundPosition: '-700px 0',
          },
          '100%': {
            backgroundPosition: '700px 0',
          },
        },
        raindrop: {
          '0%': { transform: 'translateY(-128px)' },
          '100%': { transform: 'translateY(calc(100vh + 120px)' },
        },
      },
      animation: {
        flicker: 'flicker 3s linear infinite',
        shimmer: 'shimmer 1.3s linear infinite',
        raindrop: 'animate-raindrop 5s linear infinite',
      },
      borderWidth: {
        '1': '1px',
      },
    },
  },
  plugins: [require('@tailwindcss/forms')],
} satisfies Config;
