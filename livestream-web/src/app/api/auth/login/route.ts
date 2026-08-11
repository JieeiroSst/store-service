import { NextRequest, NextResponse } from "next/server";
import { login } from "@/lib/userServiceClient";
import { SESSION_COOKIE } from "@/lib/session";

export async function POST(req: NextRequest) {
  const { username, password } = (await req.json()) as { username?: string; password?: string };
  if (!username || !password) {
    return NextResponse.json({ error: "username and password are required" }, { status: 400 });
  }

  const result = await login(username, password);
  if (!result) {
    return NextResponse.json({ error: "invalid username or password" }, { status: 401 });
  }

  const res = NextResponse.json({ ok: true });
  // httpOnly: the JWT never touches client-side JS, only this server and
  // livestream_service (which verifies its signature) ever see it.
  res.cookies.set(SESSION_COOKIE, result.sessionToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: result.expirySeconds,
  });
  return res;
}
