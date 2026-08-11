const BASE_URL = process.env.USER_SERVICE_URL ?? "http://localhost:1235";

export interface LoginResult {
  sessionToken: string;
  refreshToken: string;
  expirySeconds: number;
}

export async function login(username: string, password: string): Promise<LoginResult | null> {
  const res = await fetch(`${BASE_URL}/api/v1/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
    cache: "no-store",
  });
  if (!res.ok) return null;

  const body = (await res.json()) as Record<string, unknown>;
  const sessionToken = (body.sessionToken ?? body.session_token) as string | undefined;
  const refreshToken = (body.refreshToken ?? body.refresh_token) as string | undefined;
  const expiryTime = (body.expiryTime ?? body.expiry_time) as number | string | undefined;
  if (!sessionToken) return null;

  return {
    sessionToken,
    refreshToken: refreshToken ?? "",
    expirySeconds: Number(expiryTime) || 24 * 60 * 60,
  };
}
