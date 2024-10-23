'use client'
import Sidebar from "@/components/Sidebar";
import MainContainer from "@/components/MainContainer";
import { useEffect, useState } from 'react';
import { useUser } from '@auth0/nextjs-auth0/client';
import { getAccessToken } from '@auth0/nextjs-auth0';
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
      console.log('fetchUserData');
      console.log(user);
      if (user) {
        try {
          // Get the access token from Auth0
          // todo double check this and see if this is the most secure way of making auth0 api calls 
          const response = await fetch('/api/user');
          console.log(response);
          if (!response.ok) {
            throw new Error(
              `Failed to fetch user data: ${response.status} ${response.statusText}`
            );
          }

          const data = await response.json();
          setUserData(data);
        } catch (err) {
          console.error(err);
          setFetchError(err instanceof Error ? err.message : 'An error occurred');
        }
      }
    };
    fetchUserData();
  }, [user]);

  useEffect(() => {
    const fetchUserData = async () => {
      console.log('fetchUserData');
      if (user) {
        
        try {
          const session = await getAccessToken();
          console.log("access token");
          // Get the access token from Auth0
          // todo double check this and see if this is the most secure way of making auth0 api calls 
          const response = await fetch('http://localhost:8000/api/user', {
            headers: {
              'Authorization': `Bearer ${session?.accessToken}`,
              'Content-Type': 'application/json',
            },
            credentials: 'include', // Include if you need cookies
          });
          console.log(response);
          if (!response.ok) {
            throw new Error(
              `Failed to fetch user data: ${response.status} ${response.statusText}`
            );
          }

          const data = await response.json();
          setUserData(data);
        } catch (err) {
          console.error(err);
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
