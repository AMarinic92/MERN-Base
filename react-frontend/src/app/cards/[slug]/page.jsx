'use client';

import { use, useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import API from '@/lib/api';
import ManaCostDisplay from '@/components/card/manaCostDisplay';
import Display from '@/components/card/display';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import OracleText from '@/components/card/oracleText';
import Image from 'next/image';
import { Button } from '@/components/ui/button';
import Loading from '@/components/loading/loading';

export default function CardPage({ params }) {
  const { slug } = use(params); // Make sure slug is properly unwrapped
  const [similar, setSimilar] = useState([]);
  const [variants, setVariants] = useState([]);
  const { data, isLoading, error } = useQuery({
    queryKey: ['card-info', slug],
    queryFn: async () => {
      const response = await API.get('/cards/id', {
        id: slug,
      });
      return response.data ?? {};
    },
    enabled: !!slug, // Add this - only run when slug exists
  });

  // Similar Cards Mutation
  const similarMutation = useMutation({
    mutationFn: async (currentCard) => {
      const payload = {
        oracle_id: currentCard?.OracleID,
      };
      console.debug(payload);
      return await API.post('/cards/mems', payload);
    },
    onSuccess: (data) => setSimilar(data),
  });

  const variantMutation = useMutation({
    mutationFn: async (currentCard) => {
      const payload = {
        oracle_id: currentCard?.OracleID,
        id: currentCard?.ID,
      };
      console.debug(payload);
      return await API.post('/cards/variants', payload);
    },
    onSuccess: (data) => setVariants(data),
  });
  const cards = useMemo(() => {
    if (!data) return undefined;
    const images = JSON.parse(data?.ImageURIs);
    const out = !!data?.CardFaces
      ? JSON.parse(data?.CardFaces)
      : [{ ...data, ImageURIs: images }];
    return out;
  }, [data]);

  useEffect(() => {
    if (!cards) return;
    variantMutation.mutate(cards?.[0]);
    similarMutation.mutate(cards?.[0]);
  }, [cards]);

  const filterType = cards?.map((c) => {
    const s = c.TypeLine.split(' — ');
    return s[0] ?? '';
  });

  const colorIdentity = cards?.reduce(
    (acc, curr) => acc.concat(curr?.ColorIdentity),
    [],
  );
  if (!slug) return <div>No slug provided</div>;
  if (isLoading) return <Loading />;
  if (error) return <div>Error: {error.message}</div>;
  console.log(variantMutation?.data);
  return (
    <div className="flex flex-col h-fit">
      {cards?.map((card, i) => {
        return (
          <div key={`${card.ID}-${i}`} className="flex flex-row w-fit">
            <div className="flex-col items-left p-4 m-4 gap-4 border rounded-2xl bg-card">
              <div
                className="flex flex-col gap-4 items-center"
                style={{ width: '400px' }}
              >
                <Image
                  src={card?.ImageURIs?.png || card?.iamge_uris?.png}
                  width={488}
                  height={680}
                  className="h-auto w-full rounded-[4.75%] shadow-2xl"
                  alt={card?.Name}
                  priority={true}
                />
              </div>
            </div>
            <div>
              <div className="flex flex-row text-4xl min-w-full items-left p-4 m-4 flex-nowrap gap-4 border rounded-2xl bg-card">
                {card?.Name}
                <ManaCostDisplay manaCost={card?.ManaCost} />
              </div>
              <div className="flex flex-col text-4xl min-w-fill items-left p-4 m-4 gap-4 border rounded-2xl bg-card whitespace-pre-line">
                {card?.TypeLine}
              </div>

              <div className="flex flex-col text-4xl min-w-fill items-left p-4 m-4 gap-4 border rounded-2xl bg-card whitespace-pre-line">
                <OracleText text={data?.OracleText} size="lg" />
              </div>
              <div
                className={`${card?.Power && card?.Toughness ? '' : 'hidden'} flex flex-col text-4xl min-w-fill items-left p-4 m-4 gap-4 border rounded-2xl bg-card whitespace-pre-line`}
              >
                {`${card?.Power ?? 0}/${card?.Toughness ?? 0}`}
              </div>
            </div>
          </div>
        );
      })}
      <Accordion type="multiple" collapsible>
        <AccordionItem value="similar">
          <AccordionTrigger className="text-2xl">
            Similar Cards
          </AccordionTrigger>
          <AccordionContent>
            <Display
              cards={similarMutation?.data}
              isLoading={isLoading || similarMutation.isPending}
              filterType={filterType}
              colorIdentity={colorIdentity}
            />
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="variants">
          <AccordionTrigger className="text-2xl">
            Card Variants
          </AccordionTrigger>
          <AccordionContent>
            <Display
              cards={variantMutation?.data}
              isLoading={isLoading || variantMutation.isPending}
              filterType={filterType}
              colorIdentity={colorIdentity}
            />
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  );
}
