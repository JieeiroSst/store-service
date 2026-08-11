"use client";

import { useEffect, useState } from "react";
import { getOrCreateViewerSessionId, getViewerCount, viewerHeartbeat } from "@/lib/apiClient";

const HEARTBEAT_INTERVAL_MS = 15_000;

export default function ViewerCount({ roomId, initialCount }: { roomId: string; initialCount: number }) {
  const [count, setCount] = useState(initialCount);

  useEffect(() => {
    const sessionId = getOrCreateViewerSessionId();

    const beat = () => {
      void viewerHeartbeat(roomId, sessionId);
      void getViewerCount(roomId).then(setCount);
    };
    beat();
    const interval = setInterval(beat, HEARTBEAT_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [roomId]);

  return (
    <span className="inline-flex items-center gap-1.5 rounded-full bg-neutral-900 px-2.5 py-1 text-xs font-medium text-neutral-300 ring-1 ring-white/10">
      <span className="h-1.5 w-1.5 rounded-full bg-red-500" />
      {count.toLocaleString()} {count === 1 ? "viewer" : "viewers"}
    </span>
  );
}
