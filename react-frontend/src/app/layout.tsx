import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"
import "@/app/globals.css"  // or "./globals.css" depending on your structure
import Image from "next/image"

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>

        <SidebarProvider>
            <AppSidebar />
            <main className="flex-1">
              <div  className=" flex flex-row gap-x-2.5 border-8 border-red-500 w-full">
                <SidebarTrigger />
                <Image
                  className="dark:invert border-amber-100 border-2"
                  src="/next.svg"
                  alt="Next.js logo"
                  width={100}
                  height={10}
                  priority
                />
                
              </div>
              <div className="border-8 border-blue-500 p-4">
                {children}
              </div>
            </main>
        </SidebarProvider>
      </body>
    </html>
  )
}