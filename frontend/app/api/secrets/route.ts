// app/api/secrets/route.ts
import { SecretsManagerClient, GetSecretValueCommand } from "@aws-sdk/client-secrets-manager";
import { NextResponse } from 'next/server';

// Cache at the API route level
let secretsCache: Record<string, string> | null = null;
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
let lastFetchTime = 0;

export async function GET() {
    console.log("here")
  // Check cache freshness
  if (secretsCache && (Date.now() - lastFetchTime) < CACHE_DURATION) {
    return NextResponse.json(secretsCache);
  }

  const client = new SecretsManagerClient({ 
    region: process.env.AWS_REGION,
    credentials: {
      accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
      secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!
    }
  });

  try {
    const response = await client.send(
      new GetSecretValueCommand({
        SecretId: `traba-${process.env.ENV}-frontend-config`,
      })
    );

    const secrets = JSON.parse(response.SecretString || '{}');
    // Set environment variables
    Object.entries(secrets).forEach(([key, value]) => {
    process.env[key] = value as string;
    });

    console.log('api: ', process.env.AUTH0_SECRET);

    
    // Update cache
    secretsCache = secrets;
    lastFetchTime = Date.now();
    // const secrets = JSON.parse(response.SecretString || '{}');
    
    // Return all secrets in a structured way
    return NextResponse.json({
      secrets: secrets,  // This includes all your secrets
      timestamp: Date.now()
    });
  } catch (error) {
    console.error('Error fetching secrets:', error);
    return NextResponse.json(
      { error: 'Failed to fetch configuration' },
      { status: 500 }
    );
  }
}