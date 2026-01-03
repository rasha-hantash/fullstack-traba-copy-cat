import { NextResponse } from "next/server";

export const runtime = "nodejs"; // safe default on ECS

export async function GET() {
  return NextResponse.json({ status: "ok" }, { status: 200 });
}
