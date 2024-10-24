import { NextResponse } from 'next/server'
import { verifyAndDecodeSessionToken } from '@/utils/auth' // Adjust import path as needed

// Export named function for POST method
export async function POST(request: Request) {
  try {
    // Parse the request body
    const body = await request.json()
    const { token } = body

    if (!token) {
      return NextResponse.json(
        { error: 'Token is required' },
        { status: 400 }
      )
    }

    const decodedToken = await verifyAndDecodeSessionToken(token)
    console.log("decodedToken", decodedToken)
    
    return NextResponse.json(decodedToken, { status: 200 })
    
  } catch (error) {
    console.log("error", error)
    return NextResponse.json(
      { error: error instanceof Error ? error.message : 'Unknown error occurred' },
      { status: 401 }
    )
  }
}