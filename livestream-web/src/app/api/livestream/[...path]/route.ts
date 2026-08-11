import { NextRequest, NextResponse } from "next/server";
import { getRawSessionToken } from "@/lib/session";

const BASE_URL = process.env.LIVESTREAM_SERVICE_URL ?? "http://localhost:8080";

async function proxy(req: NextRequest, path: string[]): Promise<NextResponse> {
  const target = `${BASE_URL}/api/${path.join("/")}${req.nextUrl.search}`;
  const token = getRawSessionToken();

  const headers: Record<string, string> = {};
  if (token) headers.Authorization = `Bearer ${token}`;

  const hasBody = req.method !== "GET" && req.method !== "DELETE" && req.method !== "HEAD";
  if (hasBody) headers["Content-Type"] = "application/json";

  const upstream = await fetch(target, {
    method: req.method,
    headers,
    body: hasBody ? await req.text() : undefined,
    cache: "no-store",
  });

  const text = await upstream.text();
  return new NextResponse(text || null, {
    status: upstream.status,
    headers: { "Content-Type": upstream.headers.get("Content-Type") ?? "application/json" },
  });
}

export async function GET(req: NextRequest, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path);
}
export async function POST(req: NextRequest, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path);
}
export async function DELETE(req: NextRequest, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path);
}
