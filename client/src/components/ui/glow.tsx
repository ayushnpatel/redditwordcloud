import * as React from 'react';
import '@/styles/colors.css';
import { cn } from '@/lib/utils';

type GlowProps = React.ComponentProps<'div'>;
export default function Glow({ children, className }: GlowProps) {
  return (
    <div className='relative'>
      <div
        className={cn(
          'absolute -inset-0.5 bg-gradient-to-r  from-cyan-300 to-purple-600 rounded-lg blur-sm opacity-75 group-hover:opacity-100 transition duration-1000 group-hover:duration-200 '
        )}
      ></div>
      <div className='relative'>{children}</div>
    </div>
  );
}
