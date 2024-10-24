'use client';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation'
import EmailVerification from '@/components/EmailVerification';

export default function VerifyEmailPage() {
  const searchParams = useSearchParams()
 
  const session_token  = searchParams.get('session_token')
  const [email, setEmail] = useState<string | null>(null);
  const [auth0UserId, setAuth0UserId] = useState<string | null>(null);
  // const [userData, setUserData] = useState<{ name: string | null }>({ name: null });
  const [error, setError] = useState<string | null>(null);

  // const { user, error, isLoading } = useUser();


  useEffect(() => {

   
    
    const verifyToken = async () => {
      if (typeof session_token !== 'string') {
        setError('No token provided'); // todo do i need these setError? 
        // setIsVerifying(false);
        return;
      }

      try {
        // First decode to show the email immediately (optional)
        // const decoded = jwtDecode<TokenPayload>(session_token);
        // setEmail(decoded.email);

        // Then verify the signature server-side
        console.log("session_token", session_token);
        const response = await fetch('http://localhost:3000/api/verify-token', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ token: session_token }),
        });
        console.log("response", response);

        if (!response.ok) {
          throw new Error('Token verification failed');
        }

        const verified = await response.json();
        setEmail(verified.email);
        setAuth0UserId(verified.auth0UserId);
      } catch (err) {
        console.log("this is the error", err)
        setError('Invalid or expired token');
        setEmail(null);
      }
    };

    verifyToken();
  }, [session_token]);

  // if (isLoading) return <div>Loading...</div>;
  if (error) return <div>{error}</div>;

  // if (!user) return null; // This prevents the main content from flashing before redirect


  // useEffect(() => {
  //   const fetchUserData = async () => {
  //     try {
  //       const response = await fetch('/api/my-api');
  //       if (!response.ok) {
  //         throw new Error('Failed to fetch user data');
  //       }
  //       const data = await response.json();
  //       setUserData(data);
  //     } catch (err) {
  //       setError(err instanceof Error ? err.message : 'An error occurred');
  //     }
  //   };

  //   fetchUserData();
  // }, []);

  if (error) {
    return <div>Error: {error}</div>;
  }


  return (
  
      <EmailVerification 
      email={ email || '' }
    />
    
  );
}
