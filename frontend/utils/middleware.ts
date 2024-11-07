
import { getAwsSecrets } from './getSecrets';

let secretsCache: Record<string, string> | null = null;

function getSecretsCache() {
  return secretsCache;
}

function setSecretsCache(secrets: Record<string, string>) {
  secretsCache = secrets;
}


export async function initializeEnvironment() {
  // Check if we need to fetch new secrets
  if (!process.env.AUTH0_SECRET || !getSecretsCache()) {
    try {
      const secrets = await getAwsSecrets();
      setSecretsCache(secrets);
      
      // Set environment variables
      Object.entries(secrets).forEach(([key, value]) => {
        process.env[key] = value as string;
      });
    } catch (error) {
      console.error('Failed to initialize environment variables:', error);
      throw error;
    }
  } else {
    // Restore from cache
    const cachedSecrets = getSecretsCache();
    Object.entries(cachedSecrets!).forEach(([key, value]) => {
      process.env[key] = value;
    });
  }
}