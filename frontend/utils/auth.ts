// import { jwtDecode } from 'jwt-decode';
import { verify } from 'jsonwebtoken';

// utils/auth0.js
import { initAuth0 } from '@auth0/nextjs-auth0';
import { initializeEnvironment } from './middleware';

// export default initAuth0({
//   secret: process.env.AUTH0_SECRET,
//   issuerBaseURL: process.env.AUTH0_ISSUER_BASE_URL,
//   baseURL: process.env.AUTH0_BASE_URL,
//   clientID: process.env.AUTH0_CLIENT_ID,
//   clientSecret: process.env.AUTH0_CLIENT_SECRET,
//   routes: {
//       callback: '/api/auth/callback',
//       postLogoutRedirect: '/',
//   },
//   session: {
//       absoluteDuration: 24 * 60 * 60, // 24 hours
//   },
// });


const config = async () => {
  // await initializeEnvironment();
  await initializeEnvironment();

  return {
    secret: process.env.AUTH0_SECRET,
    issuerBaseURL: process.env.AUTH0_ISSUER_BASE_URL,
    baseURL: process.env.AUTH0_BASE_URL,
    clientID: process.env.AUTH0_CLIENT_ID,
    clientSecret: process.env.AUTH0_CLIENT_SECRET,
  }
}

export default initAuth0(await config());

interface TokenPayload {
  sub: string;
  email: string;
  provider: string;
}


export interface VerificationEmailResponse {
    status: string;
    type: string;
    created_at: string;
    id: string;
  }
  
 export  interface VerificationEmailPayload {
    user_id: string;
    client_id?: string;
    identity: {
      user_id: string;
      provider: string;
    };
    organization_id?: string; // todo include organization id to make organization parameters + branding available to the email template
  }

export const verifyAndDecodeSessionToken = async (token: string): Promise<TokenPayload> => {
  try {
    // Get the secret from environment variables
    const secret = process.env.MY_REDIRECT_SECRET;
    if (!secret) {
      throw new Error('Missing secret key');
    }

    // Verify the token with the secret
    const decodedToken = verify(token, secret) as TokenPayload;
    return decodedToken;
    
  } catch (error) {
    if (error instanceof Error) {
      throw new Error(`Token verification failed: ${error.message}`);
    }
    throw new Error('Token verification failed');
  }
};