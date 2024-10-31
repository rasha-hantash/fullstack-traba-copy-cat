import { getManagementToken } from '@/utils/management-token';
import { NextResponse } from 'next/server';
import { VerificationEmailPayload } from '@/utils/auth';

export async function POST(request: Request) {
  try {
    const payload: VerificationEmailPayload = await request.json();

    if (!payload.user_id) {
      return NextResponse.json({ error: 'User ID is required' }, { status: 400 });
    }


    if (!payload.identity?.provider) {
      return NextResponse.json({ error: 'Identity provider is required' }, { status: 400 });
    }

    // Get a fresh management token
    const managementToken = await getManagementToken();

    const domain = process.env.AUTH0_ISSUER_BASE_URL;
    const response = await fetch(`${domain}/api/v2/jobs/verification-email`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${managementToken}`,
      },
      body: JSON.stringify({
        user_id: payload.user_id,
        client_id: process.env.AUTH0_CLIENT_ID,
        identity: {
          user_id: payload.identity.user_id,
          provider: payload.identity.provider,
        }
      })
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message }, { status: response.status });
    }

    const result = await response.json();
    return NextResponse.json(result);
  } catch (error) {
    console.error('Error resending verification email:', error);
    return NextResponse.json(
      { error: 'Failed to resend verification email' },
      { status: 500 }
    );
  }
}