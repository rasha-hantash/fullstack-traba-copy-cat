'use client';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation'
import EmailVerification from '@/components/EmailVerification';

export default function VerifyEmailPage() {
  const searchParams = useSearchParams()
  const session_token  = searchParams.get('session_token')
  const [email, setEmail] = useState<string | null>(null);
  const [auth0UserId, setAuth0UserId] = useState<string | null>(null);
  const [identityAuth0UserId, setIdentityAuth0UserId] = useState<string | null>(null);
  const [provider, setProvider] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const verifyToken = async () => {
      if (typeof session_token !== 'string') {
        setError('No token provided'); // todo do i need these setError? 
        return;
      }

      try {
        const response = await fetch('http://localhost:3000/api/verify-token', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ token: session_token }),
        });

        if (!response.ok) {
          throw new Error('Token verification failed');
        }

        const verified = await response.json();
        setEmail(verified.email);
        setAuth0UserId(verified.sub);
        setIdentityAuth0UserId(verified.identity.user_id);
        setProvider(verified.identity.provider);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
        setEmail(null);
      }
    };

    verifyToken();
  }, [session_token]);

  if (error) return <div>{error}</div>;

  if (error) {
    return <div>Error: {error}</div>;
  }


  return (
      <EmailVerification 
      email={ email || '' }
      auth0UserId={ auth0UserId || '' }
      identityAuth0UserId={ identityAuth0UserId || '' } 
      provider={ provider || '' }
    />
  );
}
