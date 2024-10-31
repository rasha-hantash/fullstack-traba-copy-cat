import auth from '@/utils/auth';
// import { NextRequest } from 'next/server';

// Initialize Auth0 instance
// const auth0Promise = initializeAuth0();


export const GET = auth.handleAuth({
  login: auth.handleLogin({
    authorizationParams: {
      audience: 'https://traba-api/',
      // Add the `offline_access` scope to also get a Refresh Token
      scope: 'openid profile email ' // or AUTH0_SCOPE
    }
  })
});
