"use client"
import { Button } from "@/components/ui/button";
import { useQuery } from "@tanstack/react-query";
import Image from "next/image";
import API from "../../lib/api"
import { useState } from "react";
import MtgCard from "@/components/card/mtgCard"
export default function Inbox() {
  const [getCard, setGetCard] = useState(false);
    const [getSimilar, setGetSimilar] = useState(false);

  const {data} = useQuery({ queryKey: ['rand-card'], queryFn: async () =>{
      const response = await API.get("/cards/rand");
      setGetCard(false);
      return response
  }, enabled: getCard });

    const {data: similar} = useQuery({ queryKey: ['rand-card'], queryFn: async () =>{
      const response = await API.post("/cards/similar", data?.OracleText.split(","));
      setGetCard(false);
      setGetSimilar(false)
      return response
  }, enabled: data && getSimilar});
  console.log(similar)
  return (
    <div className="flex min-h-screen items-center justify-center  font-sans dark:bg-black">
      <main className="flex min-h-screen w-full max-w-3xl flex-col items-center justify-between py-32 px-16 bg-white dark:bg-black sm:items-start">
        <div>
        <h1>Random</h1>
        <Button onClick={() =>setGetCard(true)}>Get Random</Button>
        <MtgCard data={data?.card}/>
        <Button className={`${data ? '' : 'hidden'}`}onClick={() =>setGetSimilar(true)}>Get Similar</Button>
        </div>
        
      </main>
    </div>
  );
}
