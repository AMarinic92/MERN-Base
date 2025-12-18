'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useQuery } from '@tanstack/react-query';
import { useDeferredValue, useEffect, useState } from 'react';
import z from 'zod';

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
import API from '@/lib/api';
import Loading from '@/components/loading/loading';
import MtgCard from '@/components/card/mtgCard';

const formSchema = z.object({
  cardName: z.string().max(50),
});

export default function CardSearchForm() {
  const searchForm = useForm({
    resolver: zodResolver(formSchema),
    defaultValues: { cardName: '' },
  });

  // 1. Watch the input value
  const cardName = searchForm.watch('cardName');

  // 2. Debounce the value (React 18+ hook)
  // This prevents the query from firing on every single keystroke
  const deferredTerm = useDeferredValue(cardName);

  // 3. Simple Query: It automatically refetches when deferredTerm changes
  const { data, isFetching } = useQuery({
    queryKey: ['fuzzy-card', deferredTerm],
    queryFn: async () => {
      const response = await API.get('/cards/fuzzy', {
        name: deferredTerm,
      });
      return response.data;
    },
    // Only run query if we have at least 2 characters
    enabled: deferredTerm.length >= 2,
  });
  console.log(data);
  return (
    <div className="flex flex-col items-center w-full gap-8">
      <Form {...searchForm}>
        <form className="w-full max-w-md">
          <FormField
            control={searchForm.control}
            name="cardName"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Card Name</FormLabel>
                <FormControl>
                  <Input placeholder="Black Lotus..." {...field} />
                </FormControl>
                <FormDescription>Fuzzy search our database.</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </form>
      </Form>

      <div className="flex flex-wrap justify-center gap-4">
        {isFetching ? (
          <Loading />
        ) : (
          data?.map((card) => (
            <div key={card.id} className="flex-col">
              <MtgCard key={card.id || card.Name} data={card} />
            </div>
          ))
        )}
      </div>
    </div>
  );
}
