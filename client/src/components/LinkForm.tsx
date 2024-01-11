'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';

import * as z from 'zod';

import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { toast } from '@/components/ui/use-toast';
import { makeRequest } from '@/app/actions/makereq';
import Glow from '@/components/ui/glow';
import { useState } from 'react';

const FormSchema = z.object({
  link: z.string().url({
    message: 'Link must be valid.',
  }),
});

export default function LinkForm() {
  const router = useRouter();
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      link: '',
    },
  });

  async function onSubmit(data: z.infer<typeof FormSchema>) {
    const res = await makeRequest();
    router.push(`/${res.Link}`);
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className='w-3/4 md:w-1/2 mt-6 mb-6'
      >
        <FormField
          control={form.control}
          name='link'
          render={({ field }) => {
            const [isClicked, setIsClicked] = useState(false);
            return (
              <FormItem>
                <FormControl>
                  <Glow isClicked={isClicked}>
                    <Input
                      className='bg-white dark:bg-rwc-violet-950 to-rwc-violet-700 from-rwc-cyan-700 dark:to-rwc-violet-100 dark:from-rwc-cyan-100 hover:scale-110 transition-all'
                      autoComplete='off'
                      placeholder='Insert your reddit thread link here...'
                      onClick={() => setIsClicked(true)}
                      onBlurCapture={() => setIsClicked(false)}
                      {...field}
                    />
                  </Glow>
                </FormControl>
                <FormMessage />
              </FormItem>
            );
          }}
        />
        <Button
          className='my-4 hover:scale-105 transition-all border-2 dark:border-rwc-beige-900 border-rwc-beige-50 bg-rwc-cyan-700 dark:bg-rwc-cyan-100 text-rwc-cyan-50 dark:text-rwc-cyan-700'
          type='submit'
        >
          Get Word Cloud
        </Button>
      </form>
    </Form>
  );
}
