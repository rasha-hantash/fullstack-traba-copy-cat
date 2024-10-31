// middleware.ts (at root level)
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { initializeEnvironment } from './utils/middleware';

export async function middleware(request: NextRequest) {    
    try {
        await initializeEnvironment();
    } catch (error) {
        console.error('Error in middleware:', error);
    }
    
    return NextResponse.next();
}

// export const config = {
//   matcher: [
//     // Make sure this exactly matches your auth route pattern
//     '/api/auth/:path*',
//     '/api/auth/[...auth0]/:path*',
//     '/api/auth/[...auth0]', // Add this to catch the exact route
//   ]
// }