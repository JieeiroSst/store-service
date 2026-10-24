import { cookies } from "next/headers";
import { decodeJwtPayload, isExpired } from "@/lib/jwt";
import type { Session } from "@/lib/types";

export const SESSION_COOKIE = "session_token";

export async function getRawSessionToken(): Promise<string | undefined> {
  return (await cookies()).get(SESSION_COOKIE)?.value;
}

export async function getSession(): Promise<Session | null> {
  const token = await getRawSessionToken();
  if (!token) return null;

  const payload = decodeJwtPayload(token);
  if (!payload?.sub || isExpired(payload)) return null;

  return {
    userId: payload.sub,
    username: payload.username ?? "",
    role: payload.role ?? "",
  };
}
