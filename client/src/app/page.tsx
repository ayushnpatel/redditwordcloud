'use server';

import Head from 'next/head';
import * as React from 'react';

import ArrowLink from '@/components/links/ArrowLink';
import ButtonLink from '@/components/links/ButtonLink';
import UnderlineLink from '@/components/links/UnderlineLink';
import UnstyledLink from '@/components/links/UnstyledLink';

import LinkForm from '@/components/LinkForm';

/**
 * SVGR Support
 * Caveat: No React Props Type.
 *
 * You can override the next-env if the type is important to you
 * @see https://stackoverflow.com/questions/68103844/how-to-override-next-js-svg-module-declaration
 */
import Logo from '~/svg/redditwordclouddark.svg';
import { Label } from '@/components/ui/label';
import Glow from '@/components/ui/glow';
import { Button } from '@/components/ui/button';
import { Header } from '@/components/Header';
import dynamic from 'next/dynamic';
import RedditWordCloudLogo from '@/components/RedditWordCloudLogo';

const DynamicHeader = dynamic(
  () => import('../components/RaindropBackground'),
  {
    ssr: true,
  }
);

export default async function HomePage() {
  return (
    <main>
      <Head>
        <script></script>
        <title>redditwordcloud</title>
      </Head>
      <section className='bg-white dark:bg-rwc-violet-950'>
        <DynamicHeader></DynamicHeader>
        <div className='layout relative flex min-h-screen flex-col items-center justify-center py-12 text-center'>
          <Header />
          <RedditWordCloudLogo className='w-24 sm:w-32 md:w-48 lg:w-64 -m-6 ' />
          <h1 className='mt-4 text-3xl font-semibold text-rwc-beige-900 dark:text-rwc-beige-50'>
            redditwordcloud
          </h1>
          <LinkForm />

          <footer className='absolute bottom-2 text-rwc-beige-900 dark:text-rwc-beige-50'>
            © {new Date().getFullYear()} By{' '}
            <UnderlineLink href='https://linkedin.com/in/ayushnpatel'>
              Ayush Patel
            </UnderlineLink>
          </footer>
        </div>
      </section>
    </main>
  );
}
