// app/api/user/route.ts
import auth from '@/utils/auth';
import { config } from '../config';
import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';

export async function GET() {
  try {
    const token = await auth.getAccessToken();
    console.log('apiUrl', config.apiUrl)
    const response = await fetch(`${config.apiUrl}/api/user`, {
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