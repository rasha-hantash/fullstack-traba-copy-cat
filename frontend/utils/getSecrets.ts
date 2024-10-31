import { SecretsManagerClient, GetSecretValueCommand } from "@aws-sdk/client-secrets-manager";

export async function getAwsSecrets() {
  let environment 
    if (process.env.NODE_ENV === 'development') {
        environment = 'local'
    } else {
        environment = process.env.NODE_ENV
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
        SecretId: `local-frontend-config`,
      })
    );;

    return JSON.parse(response.SecretString || '{}');
  } catch (error) {
    console.error('Error fetching secrets:', error);
    throw error;
  }
}
