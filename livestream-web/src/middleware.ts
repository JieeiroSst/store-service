import { NextRequest, NextResponse } from "next/server";
import { decodeJwtPayload, isExpired } from "@/lib/jwt";
import { SESSION_COOKIE } from "@/lib/session";

export function middleware(req: NextRequest) {
  const token = req.cookies.get(SESSION_COOKIE)?.value;
  const payload = token ? decodeJwtPayload(token) : null;
  const authenticated = Boolean(payload?.sub) && !isExpired(payload ?? {});

  if (!authenticated) {
    const loginUrl = new URL("/login", req.url);
    loginUrl.searchParams.set("next", req.nextUrl.pathname);
    return NextResponse.redirect(loginUrl);
  }

  if (req.nextUrl.pathname.startsWith("/admin") && payload?.role !== "admin") {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/admin/:path*"],
};
