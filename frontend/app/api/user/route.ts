// app/api/user/route.ts
import { getAccessToken  } from '@auth0/nextjs-auth0';
import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';

export async function GET() {
  try {
    const token = await getAccessToken();

    const response = await fetch('http://localhost:8000/api/user', {
      credentials: 'include',  // Add this line
      headers: {
        'Authorization': `Bearer ${token.accessToken}`,
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
      { error: `Failed to fetch user data` }, 
      { status: 500 }
    );
  }
}