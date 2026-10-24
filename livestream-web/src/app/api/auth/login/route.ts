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
  // httpOnly: the cookie itself is never readable from client-side JS. The
  // room page hands the token to the chat socket only for signed-in users,
  // because browsers cannot attach an Authorization header to a WebSocket.
  res.cookies.set(SESSION_COOKIE, result.sessionToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: result.expirySeconds,
  });
  return res;
}
