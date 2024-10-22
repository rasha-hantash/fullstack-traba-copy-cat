'use client'
import Sidebar from "@/components/Sidebar";
import MainContainer from "@/components/MainContainer";
import { useEffect, useState } from 'react';
import { useUser } from '@auth0/nextjs-auth0/client';
import { useRouter } from 'next/navigation';


interface UserData {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  company_name: string;
  phone_number: string;
}


export default function Home() {
  const { user, error, isLoading } = useUser();
  const router = useRouter();
  const [userData, setUserData] = useState<UserData | null>(null);
  const [fetchError, setFetchError] = useState<string | null>(null);

  useEffect(() => {
    if (!isLoading && !user) {
      router.push('/api/auth/login');
    }
  }, [isLoading, user, router]);


  useEffect(() => {
    const fetchUserData = async () => {
      if (user) {
        try {
          const response = await fetch('/api/user');
          if (!response.ok) {
            throw new Error('Failed to fetch user data');
          }
          const data = await response.json();
          setUserData(data);
        } catch (err) {
          setFetchError(err instanceof Error ? err.message : 'An error occurred');
        }
      }
    };

    fetchUserData();
  }, [user]);

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
