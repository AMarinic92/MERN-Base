import { Label } from "@/components/ui/label";
import Image from "next/image";
import { useMemo } from "react";
export default function MtgCard({data}){
    const imageUri = useMemo(() => {
        if(!data) return;
        const images = data?.ImageURIs ? JSON.parse(data?.ImageURIs) : JSON.parse(data?.CardFaces);
        
        return data?.ImageURIs ? [images?.normal] : [image?.[0]?.ImageURIs?.normal, image?.[1]?.ImageURIs.normal];
    }, [data])
    if(!data) return;
    console.log(data)
    console.log(imageUri)
    return(
        <div className={`${getStyle(data?.Colors)} min-h-96 min-w-48 text-6xl p-3.5 mt-2`}>
            <div className="m-4">
                {data?.Name}
            </div>
            <div>
            {imageUri.map((uri) =>{
                            return <Image 
                                src={uri}
                                width={500}
                                height={500}
                                preload={true}
                                alt="Picture of the author"
                            />}
            )}
            </div>
        </div>
    )
}
const colourMap = new Map([
    ["B", "bg-zinc-600 border-zinc-900"], 
    ["W", "bg-slate-50 border-slate-200"],
    ["U", "bg-zinc-600 border-zinc-900"], 
    ["G", "bg-green-700 border-green-950"],
    ["R", "bg-red-800 border-red-950"],
]);

function getStyle(colours){
    
    const border = "border-8"
    if(!colours) return "bg-gray-400 border-gray-900 "+border;
    if(colours?.length > 1){
        return "bg-yellow-600 border-yellow-900 "+border;
    }
    return `${colourMap.get(colours[0])} ${border}`;

}