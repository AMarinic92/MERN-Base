import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"
import "@/app/globals.css"  // or "./globals.css" depending on your structure

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <SidebarProvider>
          <div id='main-header'className="flex min-h-screen border-8 border-amber-400">
            <AppSidebar />
            <main className="flex-1 border-8 border-green-500">
              <div className="border-8 border-red-500">
                <SidebarTrigger />
              </div>
              <div className="border-8 bg-amber-800 border-blue-500 p-4">
                {children}
              </div>
            </main>
          </div>
        </SidebarProvider>
      </body>
    </html>
  )
}