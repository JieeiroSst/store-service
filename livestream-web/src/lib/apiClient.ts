"use client";

async function call(path: string, method: string, body?: unknown): Promise<Response> {
  return fetch(`/api/livestream/${path}`, {
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
}

export async function createRoom(title: string, description: string) {
  const res = await call("v1/rooms", "POST", { title, description });
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to create room");
  return res.json();
}

export async function regenerateStreamKey(roomId: string): Promise<string> {
  const res = await call(`v1/rooms/${roomId}/stream-key/regenerate`, "POST");
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to regenerate stream key");
  return (await res.json()).streamKey as string;
}

export async function requestIngest(roomId: string) {
  const res = await call(`v1/rooms/${roomId}/ingest`, "POST");
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to request ingest endpoint");
  return res.json();
}

export async function endStream(roomId: string) {
  const res = await call(`v1/rooms/${roomId}/end`, "POST");
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to end stream");
}

export async function banFromChat(roomId: string, targetUserId: string, durationSeconds: number) {
  const res = await call(`v1/rooms/${roomId}/chat/ban`, "POST", { targetUserId, durationSeconds });
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to ban user");
}

export async function unbanFromChat(roomId: string, targetUserId: string) {
  const res = await call(`v1/rooms/${roomId}/chat/unban`, "POST", { targetUserId });
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to unban user");
}

export async function adminDeleteRoom(roomId: string) {
  const res = await call(`v1/admin/rooms/${roomId}`, "DELETE");
  if (!res.ok) throw new Error((await res.json()).error ?? "failed to delete room");
}

export async function viewerHeartbeat(roomId: string, sessionId: string) {
  await call(`v1/rooms/${roomId}/viewers/heartbeat`, "POST", { sessionId });
}

export async function reportQoE(roomId: string, bitrateKbps: number, bufferingEvents: number) {
  await call(`v1/rooms/${roomId}/qoe`, "POST", { bitrateKbps, bufferingEvents });
}

export async function getViewerCount(roomId: string): Promise<number> {
  const res = await fetch(`/api/livestream/v1/rooms/${roomId}/viewers`, { cache: "no-store" });
  if (!res.ok) return 0;
  return ((await res.json()) as { viewers: number }).viewers;
}

export function getOrCreateViewerSessionId(): string {
  const key = "livestream_viewer_session_id";
  let id = typeof window !== "undefined" ? window.localStorage.getItem(key) : null;
  if (!id) {
    id = crypto.randomUUID();
    if (typeof window !== "undefined") window.localStorage.setItem(key, id);
  }
  return id;
}
