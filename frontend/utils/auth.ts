// import { jwtDecode } from 'jwt-decode';
import { verify } from 'jsonwebtoken';

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

    // const decoded = decode()jwtDecode<TokenPayload>(token);
    // console.log("decoded", decoded);
    // Verify the token with the secret
    const decodedToken = verify(token, secret) as TokenPayload;
    // const decoded = jwtDecode<TokenPayload>(token);
    console.log("decodedToken", decodedToken);

    return decodedToken;
  } catch (error) {
    if (error instanceof Error) {
      throw new Error(`Token verification failed: ${error.message}`);
    }
    throw new Error('Token verification failed');
  }
};