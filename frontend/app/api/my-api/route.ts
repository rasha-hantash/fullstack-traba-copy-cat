import { getSession, getAccessToken  } from '@auth0/nextjs-auth0';
import { NextResponse } from 'next/server';

const GET = async () => {
  const session = await getSession();
  const token = await getAccessToken();
  console.log("session", session);
  console.log("token", token);

  return NextResponse.json(session?.user);
};

export { GET };

// export const runtime = 'edge';