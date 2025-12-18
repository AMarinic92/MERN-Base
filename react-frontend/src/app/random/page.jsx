'use client';
import { Button } from '@/components/ui/button';
import { useMutation } from '@tanstack/react-query';
import Image from 'next/image';
import API from '../../lib/api';
import MtgCard from '@/components/card/mtgCard';
import { useEffect, useState } from 'react';
import { Spinner } from '@/components/ui/spinner';
import Loading from '@/components/loading/loading';
export default function RandomCard() {
  const [similar, setSimilar] = useState([]);
  // Random Card Mutation
  const randomMutation = useMutation({
    mutationFn: () => API.get('/cards/rand'),
  });

  // Similar Cards Mutation
  const similarMutation = useMutation({
    mutationFn: (currentCard) => {
      const payload = {
        name: currentCard?.Name,
        oracle_texts: currentCard?.OracleText?.split('\n') || [],
      };
      console.debug(payload);
      return API.post('/cards/similar', payload);
    },
  });

  useEffect(() => {
    setSimilar([]);
  }, [randomMutation.data]);
  useEffect(() => {
    console.log('sim', similarMutation.data);
    setSimilar(similarMutation.data);
  }, [similarMutation.data]);
  const handleGetRandom = () => {
    randomMutation.mutate();
  };

  const handleGetSimilar = () => {
    similarMutation.mutate(randomMutation.data?.card);
  };
  return (
    <div className="flex flex-col items-center">
      <div className="flex flex-row gap-x-3">
        <div>
          <Button onClick={handleGetRandom} disabled={randomMutation.isPending}>
            {randomMutation.isPending ? 'Loading...' : 'Get Random'}
          </Button>
        </div>
        <div>
          <Button
            onClick={handleGetSimilar}
            disabled={
              similar?.length ||
              !randomMutation.data ||
              similarMutation.isPending
            }
          >
            {similarMutation.isPending ? 'Finding...' : 'Get Similar'}
          </Button>
        </div>
      </div>

      {/* Access data via mutation.data */}
      <MtgCard
        isLoading={randomMutation.isPending}
        data={randomMutation.data?.card}
      />

      <div className={`${similarMutation.isPending ? '' : 'hidden'}`}>
        <Loading />
      </div>
      <div className="flex flex-wrap flex-row justify-center items-center w-fit h-fit max-h-fit max-w-fit">
        {similar ? (
          similar?.map((simCard) => {
            return (
              <div key={simCard.id || simCard.Name} className="flex-col">
                <MtgCard isLoading={similarMutation.isPending} data={simCard} />
              </div>
            );
          })
        ) : (
          <></>
        )}
      </div>
    </div>
  );
}
