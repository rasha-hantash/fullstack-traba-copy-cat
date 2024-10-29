// app/api/user/route.ts
import { getSession  } from '@auth0/nextjs-auth0';
import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  try {
    // Get the search query from URL
    const url = new URL(request.url);
    const searchQuery = url.searchParams.get('search') || '';
    
    const session = await getSession();

    // Construct the backend URL with search parameter
    const backendUrl = new URL('http://localhost:8000/api/invoices');
    if (searchQuery) {
      backendUrl.searchParams.set('search', searchQuery);
    }

    const response = await fetch(backendUrl.toString(), {
      credentials: 'include',
      headers: {
        'Authorization': `Bearer ${session?.accessToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Backend responded with status: ${response.status}`);
    }

    const data = await response.json();

    // Return the response with proper headers
    return new Response(JSON.stringify(data), {
      headers: {
        'Content-Type': 'application/json',
      },
      status: 200,
    });

  } catch (error) {
    console.error('API route error:', error);
    
    // Return a proper error response
    return new Response(
      JSON.stringify({
        error: error instanceof Error ? error.message : 'An unexpected error occurred',
      }),
      {
        headers: {
          'Content-Type': 'application/json',
        },
        status: error instanceof Error && error.message.includes('status: 401') ? 401 : 500,
      }
    );
  }
}