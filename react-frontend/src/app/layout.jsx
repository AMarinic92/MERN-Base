'use client';
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { AppSidebar } from '@/components/app-sidebar';
import '@/app/globals.css'; // or "./globals.css" depending on your structure
import Image from 'next/image';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';

export default function Layout({ children }) {
  const queryClient = new QueryClient();
  return (
    <html lang="en">
      <body className="dark bg-background">
        <QueryClientProvider client={queryClient}>
          <SidebarProvider>
            <AppSidebar />
            <main className="flex-1">
              <div className=" flex flex-row gap-x-2.5 w-full h-fit p-4">
                <SidebarTrigger className="mt-4" />
                <div className="text-9xl">Card Barrage</div>
              </div>
              <div className="w-full p-4">{children}</div>
            </main>
          </SidebarProvider>
        </QueryClientProvider>
      </body>
    </html>
  );
}
