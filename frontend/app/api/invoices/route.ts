// app/api/user/route.ts
import auth from '@/utils/auth';
import { config } from '../config';

export const dynamic = 'force-dynamic';

export async function GET(request: Request) {
  try {
    // Get the search query from URL
    const { searchParams } = new URL(request.url);
    const searchQuery = searchParams.get('search') || '';
    console.log('apiUrl', config.apiUrl);
    
    const token = await auth.getAccessToken()
    // Construct the backend URL with search parameter
    const backendUrl = new URL(`${config.apiUrl}/api/invoices`);
    if (searchQuery) {
      backendUrl.searchParams.set('search', searchQuery);
    }

    const response = await fetch(backendUrl.toString(), {
      credentials: 'include',
      headers: {
        'Authorization': `Bearer ${token?.accessToken}`,
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