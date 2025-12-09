import { Label } from "@/components/ui/label";
import Image from "next/image";
import { useMemo } from "react";
export default function MtgCard({data}){
    const imageUri = useMemo(() => {
        if(!data) return;
        const images =JSON.parse(data?.ImageURIs);
        return images?.normal;
    }, [data])
    if(!data) return;

    return(
        <div className={`${getStyle(data?.Colors)} min-h-96 min-w-48 text-6xl p-3.5 mt-2`}>
            <div className="m-4">
                {data?.Name}
            </div>
            <div>
                <Image 
                    src={imageUri}
                    width={500}
                    height={500}
                    alt="Picture of the author"
                />
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