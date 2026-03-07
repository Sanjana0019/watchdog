import { NextRequest, NextResponse } from "next/server";

const SERVER_URL = process.env.SERVER_URL ;

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;

  const res = await fetch(`${SERVER_URL}/events/analyse/${encodeURIComponent(id)}`);

  if (!res.ok) {
    const body = await res.text();
    return NextResponse.json(
      { error: body || "Analysis failed" },
      { status: res.status },
    );
  }

  const data = await res.json();
  return NextResponse.json(data);
}
