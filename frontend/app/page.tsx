'use client'
import Sidebar from "@/components/Sidebar";
import MainContainer from "@/components/MainContainer";
import { useEffect } from 'react';
import { useUser } from '@auth0/nextjs-auth0/client';
import { useRouter } from 'next/navigation';
export default function Home() {
  const { user, error, isLoading } = useUser();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !user) {
      router.push('/api/auth/login');
    }
  }, [isLoading, user, router]);

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>{error.message}</div>;

  if (!user) return null; // This prevents the main content from flashing before redirect



  return (
    user && (
    <div className="lg:flex">
      <div className="hidden lg:block w-60 h-screen dark:bg-black">
        <Sidebar />
      </div>
      <MainContainer />
    </div>
    )
  );
}
