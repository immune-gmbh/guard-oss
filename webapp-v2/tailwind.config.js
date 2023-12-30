// const { textColor } = require('tailwindcss/colors')
const plugin = require('tailwindcss/plugin');

module.exports = {
  mode: 'jit',
  content: ['./pages/**/*.{js,ts,jsx,tsx}', './components/**/*.{js,ts,jsx,tsx}'],
  theme: {
    fontFamily: {
      sans: ['Eurostile'],
    },
    extend: {
      colors: {
        primary: 'rgba(0, 0, 0, 0.94)',
        'purple-100': '#673355',
        'purple-200': '#8D667F',
        'purple-300': '#C6B3BF',
        'purple-400': '#54193F',
        'purple-500': '#410028',
        'purple-600': '#ECE6EA',
        'red-critical': '#FFD1D8',
        'red-notification': '#B00020',
        'red-cta': '#FF193C',
        'red-toast': '#9B0E0E',
        'slight-red': '#FDEBEC',
        'green-notification': '#238700',
        'green-light-notification': '#86C96E',
        'green-dark': '#219653',
        'green-valid-light': '#E0E9C7',
        'border-red-800': '#9b2c2c',
        'border-green-800': '#276749',
        'risk-yellow': '#FFF4D5',
        'risk-dark-yellow': '#805008',
        // 'middle-gray': '#F2F2F2', // use gray-300 from tailwind
      },
      borderRadius: {
        m: '0.25em /* 4px */',
      },
      letterSpacing: {
        super: '0.15em',
      },
      margin: {
        '-dots-icon': '-0.625rem',
      },
      backgroundImage: {
        'cell-small-solid': "url('/img/cell-small-solid.png')",
        'cell-corner': "url('/img/cell-corner.png')",
        'cell-dashboard-bg': "url('/img/cell-dashboard-bg.svg')",
        'hexagon-bg': "url('/img/hexagon_white1.svg')",
      },
      backgroundSize: {
        'cell-dashboard': 'min(100%, 1400px) auto',
      },
      spacing: {
        content: '74rem',
      },
      gridTemplateColumns: {
        'boot-status-bar': 'repeat(5,7em 1fr) 80px',
        'incidents-table': '1fr 25% min(11rem, 25%) min(3rem, 25%)',
      },
      screens: {
        '3xl': '2560px',
      },
    },
  },
  variants: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
    plugin(function ({ addUtilities }) {
      addUtilities({
        '.boot-status-hexagon': {
          'mask-image': "url('/img/boot-status-hexagon.svg')",
          'mask-repeat': 'no-repeat',
          'mask-position': 'center',
        },
      });
    }),
  ],
};
