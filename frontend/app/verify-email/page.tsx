'use client';
// app/verify-email/page.js

// import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';


// Import the EmailVerificationPage component we created earlier
// app/verify-email/page.tsx
import EmailVerification from '@/components/EmailVerification';

interface VerifyEmailPageProps {
  searchParams: {
    email: string;
    userId: string;
    returnTo: string;
  };
}

export default function VerifyEmailPage({ searchParams }: VerifyEmailPageProps) {
  const { email, userId, returnTo } = searchParams;

//   if (!email || !userId) {
//     // Handle invalid access - redirect to login or show error
//     return redirect('/login');
//   }

  return (
    <EmailVerification 
      email={email}
    //   userId={userId}
    //   returnTo={returnTo}
    />
  );
}