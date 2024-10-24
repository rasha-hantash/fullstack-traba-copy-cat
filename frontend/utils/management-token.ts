// lib/auth0/managementToken.ts
import { cache } from 'react';

interface Auth0TokenResponse {
  access_token: string;
  expires_in: number;
  scope: string;
  token_type: string;
}

class ManagementClient {
  private static instance: ManagementClient;
  private token: string | null = null;
  private tokenExpiryTime: number = 0;
  
  private constructor() {}

  static getInstance(): ManagementClient {
    if (!this.instance) {
      this.instance = new ManagementClient();
    }
    return this.instance;
  }

  async getToken(): Promise<string> {
    // Check if token exists and isn't close to expiring (buffer of 5 minutes)
    const currentTime = Date.now();
    if (this.token && this.tokenExpiryTime > currentTime + 300000) {
      return this.token;
    }

    // Token doesn't exist or is expiring soon, get a new one
    return this.refreshToken();
  }

  private async refreshToken(): Promise<string> {
    try {
      const domain = process.env.AUTH0_ISSUER_BASE_URL;
      const clientId = process.env.AUTH0_MANAGEMENT_CLIENT_ID;
      const clientSecret = process.env.AUTH0_MANAGEMENT_CLIENT_SECRET;
      const audience = `${domain}/api/v2/`;

      const response = await fetch(`${domain}/oauth/token`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          grant_type: 'client_credentials',
          client_id: clientId,
          client_secret: clientSecret,
          audience: audience
        })
      });

      if (!response.ok) {
        throw new Error('Failed to fetch management token');
      }

      const data: Auth0TokenResponse = await response.json();
      
      // Store the token and its expiry time
      this.token = data.access_token;
      // Set expiry time (converting seconds to milliseconds)
      this.tokenExpiryTime = Date.now() + (data.expires_in * 1000);

      return this.token;
    } catch (error) {
      console.error('Error refreshing management token:', error);
      throw error;
    }
  }
}

// Create a cached version of getToken for use in Route Handlers
export const getManagementToken = cache(async () => {
  const tokenManager = ManagementClient.getInstance();
  return tokenManager.getToken();
});

// app/api/resend-verification/route.ts
