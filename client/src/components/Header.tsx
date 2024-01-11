'use client';
import { Button } from '@/components/ui/button';
import Glow from '@/components/ui/glow';
import { Label } from '@/components/ui/label';
import { useTheme } from 'next-themes';
import { useState } from 'react';
import MagicWandLogo from '~/svg/MagicWand.svg';

const BetaTag = () => {
  return (
    <Glow className='-inset-0.5 from-rwc-violet-800 to-rwc-cyan-500 dark:from-rwc-violet-100 dark:to-rwc-cyan-100 rounded-3xl'>
      <div className=' dark:bg-rwc-violet-300 bg-rwc-violet-700 px-6 py-1 rounded-2xl border-indigo-100 border-2 border-opacity-30 shadow-md'>
        <Label className=' text-neutral-200 dark:text-rwc-cyan-700 '>
          Beta v1.0.0
        </Label>
      </div>
    </Glow>
  );
};

const ThemeButton = () => {
  const { theme, setTheme } = useTheme();

  return (
    <MagicWandLogo
      onClick={() => (theme == 'dark' ? setTheme('light') : setTheme('dark'))}
      className='w-10 hover:scale-110 -mt-1 transition-all dark:fill-white'
    />
  );
};

export const Header = () => {
  return (
    <header className='absolute top-4 flex justify-between text-gray-700 min-w-full px-4'>
      <BetaTag />
      <ThemeButton />
    </header>
  );
};
