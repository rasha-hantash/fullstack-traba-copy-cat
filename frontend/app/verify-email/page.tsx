'use client';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { Suspense } from 'react';
import EmailVerification from '@/components/EmailVerification';

// Separate component to handle the search params and verification logic
function VerificationContent() {
  const searchParams = useSearchParams();
  const session_token = searchParams.get('session_token');
  const [email, setEmail] = useState<string | null>(null);
  const [auth0UserId, setAuth0UserId] = useState<string | null>(null);
  const [identityAuth0UserId, setIdentityAuth0UserId] = useState<string | null>(null);
  const [provider, setProvider] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const verifyToken = async () => {
      if (typeof session_token !== 'string') {
        setError('No token provided');
        return;
      }

      try {
        const response = await fetch('/api/verify-token', {
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

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <EmailVerification 
      email={email || ''}
      auth0UserId={auth0UserId || ''}
      identityAuth0UserId={identityAuth0UserId || ''} 
      provider={provider || ''}
    />
  );
}

// Main page component with proper Suspense boundary
export default function VerifyEmailPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <VerificationContent />
    </Suspense>
  );
}