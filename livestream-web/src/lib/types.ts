export type RoomStatus = "offline" | "live";

export interface Room {
  id: string;
  ownerUserId: string;
  slug: string;
  title: string;
  description: string;
  streamKey: string;
  status: RoomStatus;
  createdAt: string;
  updatedAt: string;
}

export type StreamStatus = "pending" | "live" | "ended";

export interface Stream {
  id: string;
  roomId: string;
  nodeId: string;
  status: StreamStatus;
  startedAt?: string;
  endedAt?: string;
  peakViewer: number;
  createdAt: string;
}

export interface Recording {
  id: string;
  streamId: string;
  roomId: string;
  objectKey: string;
  durationSeconds: number;
  sizeBytes: number;
  createdAt: string;
}

export interface IngestEndpoint {
  rtmpURL: string;
  nodeId: string;
  streamKey: string;
}

export interface PlaybackInfo {
  url: string;
  expiresAt: string;
  isLive: boolean;
}

export interface ChatMessage {
  roomId: string;
  userId: string;
  username: string;
  body: string;
  sentAt: string;
}

export interface Session {
  userId: string;
  username: string;
  role: string;
}
