'use client';

import * as React from 'react';
import '@/styles/colors.css';
import { cn } from '@/lib/utils';
import { useTheme } from 'next-themes';
import RedditWordCloudDarkLogo from '~/svg/redditwordclouddark.svg';
import RedditWordCloudLightLogo from '~/svg/redditwordcloudlight.svg';

type RedditWordCloudLogoProps = React.ComponentProps<'div'>;
export default function RedditWordCloudLogo({
  className,
}: RedditWordCloudLogoProps) {
  const { systemTheme, theme, setTheme } = useTheme();
  const currentTheme = theme === 'system' ? systemTheme : theme;
  return (
    <div suppressHydrationWarning>
      {currentTheme === 'light' ? (
        <RedditWordCloudLightLogo className={cn(className)} />
      ) : (
        <RedditWordCloudDarkLogo className={cn(className)} />
      )}
    </div>
  );
}
