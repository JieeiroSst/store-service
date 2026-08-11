import type { IngestEndpoint, PlaybackInfo, Recording, Room, Stream } from "@/lib/types";

const BASE_URL = process.env.LIVESTREAM_SERVICE_URL ?? "http://localhost:8080";

async function getJSON<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${BASE_URL}${path}`, { cache: "no-store" });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

export async function listRooms(live?: boolean): Promise<Room[]> {
  const query = live ? "?live=true" : "";
  return (await getJSON<Room[]>(`/api/v1/rooms${query}`)) ?? [];
}

export async function getRoom(id: string): Promise<Room | null> {
  return getJSON<Room>(`/api/v1/rooms/${id}`);
}

export async function getActiveStream(id: string): Promise<Stream | null> {
  return getJSON<Stream>(`/api/v1/rooms/${id}/stream`);
}

export async function listRecordings(id: string): Promise<Recording[]> {
  return (await getJSON<Recording[]>(`/api/v1/rooms/${id}/recordings`)) ?? [];
}

export async function getViewerCount(id: string): Promise<number> {
  const result = await getJSON<{ viewers: number }>(`/api/v1/rooms/${id}/viewers`);
  return result?.viewers ?? 0;
}

export async function getPlaybackInfo(id: string): Promise<PlaybackInfo | null> {
  return getJSON<PlaybackInfo>(`/api/v1/rooms/${id}/playback`);
}

export async function requestIngestEndpointServer(
  id: string,
  bearerToken: string,
): Promise<IngestEndpoint | null> {
  const res = await fetch(`${BASE_URL}/api/v1/rooms/${id}/ingest`, {
    method: "POST",
    headers: { Authorization: `Bearer ${bearerToken}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return (await res.json()) as IngestEndpoint;
}
