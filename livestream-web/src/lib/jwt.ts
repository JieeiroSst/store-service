export interface DecodedJwtPayload {
  sub?: string;
  username?: string;
  role?: string;
  exp?: number;
}

export function decodeJwtPayload(token: string): DecodedJwtPayload | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  try {
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = payload.padEnd(payload.length + ((4 - (payload.length % 4)) % 4), "=");
    const json = decodeURIComponent(
      atob(padded)
        .split("")
        .map((c) => "%" + c.charCodeAt(0).toString(16).padStart(2, "0"))
        .join(""),
    );
    return JSON.parse(json) as DecodedJwtPayload;
  } catch {
    return null;
  }
}

export function isExpired(payload: DecodedJwtPayload): boolean {
  return Boolean(payload.exp && payload.exp * 1000 < Date.now());
}
