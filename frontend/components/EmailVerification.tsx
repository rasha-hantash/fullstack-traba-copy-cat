import { Mail } from 'lucide-react';
import { useState } from 'react';


interface VerificationEmailResponse {
  status: string;
  type: string;
  created_at: string;
  id: string;
}

interface VerificationEmailPayload {
  user_id: string;
  client_id?: string;
  identity: {
    user_id: string;
    provider: string;
  };
  organization_id?: string; // todo include organization id to make organization parameters + branding available to the email template
}


interface EmailVerificationPageProps {
  email: string;
  auth0UserId: string;
  identityAuth0UserId: string;
  provider: string;
}

const EmailVerificationPage = ({ email, auth0UserId, identityAuth0UserId, provider}: EmailVerificationPageProps) => {
  const [isResending, setIsResending] = useState(false);
  const [showResendSuccess, setShowResendSuccess] = useState(false);

  const handleResend = async () => {
    if (isResending) return;
    
    setIsResending(true);
    setShowResendSuccess(false);
    
    try {
      const payload: VerificationEmailPayload = {
        user_id: auth0UserId,
        identity: {
          user_id: identityAuth0UserId,
          provider: provider, // You might want to make this configurable
        },
      };

      await resendVerificationEmail(payload);
      setShowResendSuccess(true);
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'An unexpected error occurred');
    } finally {
      setIsResending(false);
    }
    
    // Hide success message after 5 seconds
    setTimeout(() => {
      setShowResendSuccess(false);
    }, 5000);
  };

  async function resendVerificationEmail(payload: VerificationEmailPayload): Promise<VerificationEmailResponse> {
    const headers = new Headers({
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    });
  
    const response = await fetch(`/api/resend-verification`, {
      method: 'POST',
      headers,
      body: JSON.stringify(payload),
      // redirect: 'follow',
    });
  
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to resend verification email');
    }
  
    return response.json();
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white border p-4 rounded-md w-full max-w-md">
        <div>
          <div className="flex justify-center mb-6">
            <div className="p-3 bg-blue-100 rounded-full">
              <Mail className="h-6 w-6 text-blue-600" />
            </div>
          </div>
          <div className="text-center text-2xl font-semibold">Verify your email</div>
          <div className="text-center mt-2">
            Complete your sign up by verifying your email address
          </div>
        </div>
        <div>
          <div className="p-3 flex items-center mb-6">
            <div className="ml-1">
              A verification email has been sent to:{' '}
              <span className="font-medium">{email}</span>
            </div>
          </div>
          
          <div className="space-y-1 text-center text-sm text-gray-600">
            <p>
              Click on the link in your email to complete your sign up.
            </p>
            <p>
              You may need to check your spam folder.
            </p>
          </div>

          <div className="mt-4 text-center">
          <span className="text-sm text-gray-600">Didn&apos;t receive the email? </span>
            <button
              onClick={handleResend}
              disabled={isResending}
              className={`text-sm font-medium ${
                isResending 
                  ? 'text-gray-400 cursor-not-allowed' 
                  : 'text-blue-600 hover:text-blue-800'
              }`}
            >
              {isResending ? 'Sending...' : "Send again"}
            </button>
            
            {showResendSuccess && (
              <p className="mt-2 text-sm text-green-600">
                Verification email sent successfully!
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmailVerificationPage;