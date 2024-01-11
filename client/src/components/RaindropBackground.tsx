'use client';

import Glow from '@/components/ui/glow';
import { useTheme } from 'next-themes';

const RaindropBackground = () => {
  const darkBgColors = [
    '#EAE6EF',
    '#D5CDDF',
    '#CBC1D7',
    '#D5ECEC',
    '#B9DFDF',
    '#9DD2D2',
    '#F4F4E1',
    '#E9E9C3',
    '#DEDEA6',
  ];

  const lightBgColors = [
    '#65507C',
    '#514064',
    '#3C304B',
    '#374151',
    '#275454',
    '1A3838',
    '#868632',
    '#4A4A1C',
    '#686827',
  ];

  const { systemTheme, theme, setTheme } = useTheme();
  const currentTheme = theme === 'system' ? systemTheme : theme;
  return (
    <div className='absolute w-screen h-screen overflow-hidden object-cover bg-cover'>
      {Array.from({ length: 35 }, (x, i) => {
        let distance = Math.random();
        let width = distance * 1 + 'px';
        let left = Math.floor(Math.random() * 105 - 2.5) + '%';
        let animationDelay = Math.random() * -20 + 's';
        let animationDuration = distance * 5 + 's';
        let background =
          'linear-gradient(transparent, ' +
          (currentTheme === 'light'
            ? lightBgColors[Math.floor(Math.random() * lightBgColors.length)]
            : darkBgColors[Math.floor(Math.random() * darkBgColors.length)]) +
          ')';
        let opacity = distance * 0.4 + 0.2;
        return (
          <div
            style={{
              borderRadius: '100% 100% 1em 1em',
              width: width,
              left: left,
              animationDelay: animationDelay,
              animationDuration: animationDuration,
              background: background,
              opacity: opacity,
            }}
            id='raindrop'
            key={i}
          ></div>
        );
      })}
    </div>
  );
};

export default RaindropBackground;
