// import { handleAuth } from '@auth0/nextjs-auth0';


// export const GET = handleAuth();

// pages/api/auth/[auth0].js
import { handleAuth, handleLogin } from '@auth0/nextjs-auth0';

export const GET = handleAuth({
  login: handleLogin({
    authorizationParams: {
      audience: 'https://traba-api/',
      // Add the `offline_access` scope to also get a Refresh Token
      scope: 'openid profile email ' // or AUTH0_SCOPE
    }
  })
});