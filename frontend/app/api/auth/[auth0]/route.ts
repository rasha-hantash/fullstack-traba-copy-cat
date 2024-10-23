import { handleAuth, handleCallback } from '@auth0/nextjs-auth0';
import { redirect } from 'next/navigation';
import { NextRequest } from 'next/server';

const afterCallback = async (
  req: NextRequest,
  session: { user: any },
  state?: { [key: string]: any }
) => {
    console.log("HELLLOOOOOOOOO")
  const { searchParams } = new URL(req.url);
  const error = searchParams.get('error');
  const errorDescription = searchParams.get('error_description');

  // todo print out session 

  console.log("session.user", session.user)

  if (error === 'access_denied' && errorDescription === 'verify your email') {
    redirect('/verify-email');
  }

  // Continue with your existing admin check if needed
//   if (session.user.isAdmin) {
//     return session;
//   } else {
    redirect('/');
//   }
};

export const GET = handleAuth({
  callback: handleCallback({ afterCallback })
});

export const POST = GET;