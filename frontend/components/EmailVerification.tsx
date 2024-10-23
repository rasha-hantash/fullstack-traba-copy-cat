import { AlertTriangle, Mail } from 'lucide-react';
// import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
// import { Alert, AlertDescription } from '@/components/ui/alert';

interface EmailVerificationPageProps {
  email: string;
}

const EmailVerificationPage = ({ email }: EmailVerificationPageProps) => {
  return (
    <div className=" min-h-screen flex items-center justify-center bg-gray-100">
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
            <AlertTriangle className="h-4 w-4" />
            {/* <AlertDescription> */}
            <div className='ml-1' >
            A verification email has been sent to:{' '}
            </div>
               
              <span className="font-medium">{email}</span>
            {/* </AlertDescription> */}
          </div>
          
          <div className="space-y-4 text-center text-sm text-gray-600">
            <p>
              Click on the link in your email to complete your sign up.
            </p>
            <p>
              You may need to check your spam folder.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmailVerificationPage;