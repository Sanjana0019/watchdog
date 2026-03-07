import { NextRequest, NextResponse } from "next/server";

const SERVER_URL = process.env.SERVER_URL ;

export async function GET(request: NextRequest) {
  const queryString = request.nextUrl.search;

  const res = await fetch(`${SERVER_URL}/events/all${queryString}`);

  if (!res.ok) {
    const body = await res.text();
    return NextResponse.json(
      { error: body || "Failed to fetch events" },
      { status: res.status },
    );
  }

  const data = await res.json();
  return NextResponse.json(data);
}
