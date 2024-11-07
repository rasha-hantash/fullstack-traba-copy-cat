// app/api/user/route.ts
// import auth from '@/utils/auth';
import { config } from '../../config';

export const dynamic = 'force-dynamic';

export async function POST(request: Request) {
  try {
    
    const body = await request.json();
    // Construct the backend URL with search parameter
    const backendUrl = new URL(`${config.apiUrl}/hook/user`);
    const response = await fetch(backendUrl.toString(), {
    method: 'POST', 
    // credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        user: body.user,
        secret: body.secret, // Make sure this env var is set
      }),
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