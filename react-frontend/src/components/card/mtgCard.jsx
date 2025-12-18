import { Label } from '@/components/ui/label';
import Image from 'next/image';
import { useMemo } from 'react';
import { Spinner } from '@/components/ui/spinner';
import Loading from '@/components/loading/loading';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';

export default function MtgCard({ data, isLoading = false }) {
  const imageUri = useMemo(() => {
    if (!data) return undefined;
    const images = JSON.parse(data?.ImageURIs);
    const image = images != null ? images?.normal : undefined;
    const cardFaces = JSON.parse(data?.CardFaces);
    const cardFacesUris =
      cardFaces != null
        ? cardFaces?.map((val) => val?.image_uris?.normal)
        : undefined;
    return image != undefined ? [image] : cardFacesUris;
  }, [data]);

  if (isLoading) return <Loading />;
  if (!data) return null;

  return (
    <Dialog>
      <DialogTrigger asChild>
        <div
          className="cursor-pointer flex flex-col items-center p-4 m-4 gap-4 border rounded-2xl hover:border-amber-500 transition-all bg-card"
          /* Change max-w-xs (20rem) to max-w-sm (24rem) or a fixed pixel value like 400px */
          style={{ width: '400px' }}
        >
          {/* Reduced text size from 6xl so the card remains the focus */}
          <h3 className="text-2xl font-bold text-center line-clamp-1">
            {data?.Name}
          </h3>

          {imageUri?.map((uri, i) => (
            <div key={`${data.Name}-${i}`} className="w-full">
              <Image
                src={uri}
                width={488}
                height={680}
                className="h-auto w-full rounded-[4.75%] shadow-2xl"
                alt={data?.Name}
                priority={true}
              />
            </div>
          ))}
        </div>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>{data?.Name}</DialogTitle>
          <DialogDescription asChild>
            <div className="flex flex-col gap-4">
              <p className="text-sm italic">{data?.OracleText}</p>
              {imageUri?.map((uri, i) => (
                <div key={`${data.Name}-dialog-${i}`} className="w-full">
                  <Image
                    src={uri}
                    width={488}
                    height={680}
                    // Apply 'h-auto' here as well
                    className="w-full h-auto rounded-[4.75%] shadow-xl"
                    alt={data.Name}
                  />
                </div>
              ))}
            </div>
          </DialogDescription>
        </DialogHeader>
      </DialogContent>
    </Dialog>
  );
}
