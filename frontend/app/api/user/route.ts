// app/api/user/route.ts
import { getSession, getAccessToken } from '@auth0/nextjs-auth0';
import { NextResponse } from 'next/server';

export async function GET() {
  try {
    const session = await getSession();
    console.log("accessToken", session?.accessToken)
    console.log("idToken", session?.idToken)



    const response = await fetch('http://localhost:8000/api/user', {
      headers: {
        'Authorization': `Bearer ${session?.idToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Backend responded with status: ${response.status}`);
    }

    const data = await response.json();
    return NextResponse.json(data);

  } catch (error) {
    console.error('API route error:', error);
    return NextResponse.json(
      { error: 'Failed to fetch user data' }, 
      { status: 500 }
    );
  }
}