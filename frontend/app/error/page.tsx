import React from 'react';
import { AlertCircle } from 'lucide-react';

const ErrorPage = () => {
  // Get error message from URL search params
//   const searchParams = new URLSearchParams(window.location.search);
//   const errorMessage = searchParams.get('message') || 'Something went wrong';

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full px-6 py-8 bg-white rounded-lg shadow-lg">
        <div className="flex flex-col items-center gap-4">
          <div className="h-16 w-16 bg-red-100 rounded-full flex items-center justify-center">
            <AlertCircle className="h-8 w-8 text-red-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Oops! An error occurred</h1>
        </div>
      </div>
    </div>
  );
};

export default ErrorPage;