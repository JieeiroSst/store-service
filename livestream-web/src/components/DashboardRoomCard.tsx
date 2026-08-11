"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { IngestEndpoint, Room } from "@/lib/types";
import { banFromChat, endStream, regenerateStreamKey, requestIngest, unbanFromChat } from "@/lib/apiClient";
import { gradientFor, initialsFor } from "@/lib/color";

function CopyableField({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div>
      <div className="mb-1 text-[11px] uppercase tracking-wide text-neutral-500">{label}</div>
      <div className="flex items-center gap-2">
        <code className="flex-1 truncate rounded bg-neutral-950 px-2 py-1.5 text-xs text-neutral-300 ring-1 ring-white/10">
          {value}
        </code>
        <button
          onClick={() => {
            void navigator.clipboard.writeText(value);
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
          }}
          className="shrink-0 rounded bg-neutral-800 px-2 py-1.5 text-xs text-neutral-300 hover:bg-neutral-700"
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
    </div>
  );
}

export default function DashboardRoomCard({ room }: { room: Room }) {
  const router = useRouter();
  const [streamKey, setStreamKey] = useState(room.streamKey);
  const [showKey, setShowKey] = useState(false);
  const [ingest, setIngest] = useState<IngestEndpoint | null>(null);
  const [banUserId, setBanUserId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function run(action: () => Promise<void>) {
    setError(null);
    setBusy(true);
    try {
      await action();
    } catch (err) {
      setError(err instanceof Error ? err.message : "action failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="overflow-hidden rounded-xl bg-neutral-900/60 ring-1 ring-white/10">
      <div className="flex items-center gap-3 border-b border-white/10 p-4">
        <div
          className="h-11 w-16 shrink-0 rounded-md"
          style={{ backgroundImage: gradientFor(room.id) }}
        />
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-neutral-800 text-xs font-semibold text-neutral-300">
          {initialsFor(room.title)}
        </span>
        <h3 className="min-w-0 flex-1 truncate font-medium">{room.title}</h3>
        {room.status === "live" ? (
          <span className="shrink-0 rounded bg-red-600 px-2 py-0.5 text-[11px] font-bold uppercase tracking-wide text-white">
            Live
          </span>
        ) : (
          <span className="shrink-0 rounded bg-neutral-800 px-2 py-0.5 text-[11px] uppercase tracking-wide text-neutral-500">
            Offline
          </span>
        )}
      </div>

      <div className="space-y-3 p-4">
        {showKey ? (
          <CopyableField label="Stream key" value={streamKey} />
        ) : (
          <div>
            <div className="mb-1 text-[11px] uppercase tracking-wide text-neutral-500">Stream key</div>
            <button
              onClick={() => setShowKey(true)}
              className="rounded bg-neutral-950 px-2 py-1.5 text-xs text-neutral-500 ring-1 ring-white/10 hover:text-neutral-300"
            >
              •••••••••••• (click to reveal)
            </button>
          </div>
        )}

        {ingest && (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <CopyableField label="RTMP URL" value={ingest.rtmpURL} />
            <CopyableField label="Stream key (for OBS)" value={ingest.streamKey} />
          </div>
        )}

        {error && <p className="text-sm text-red-400">{error}</p>}

        <div className="flex flex-wrap gap-2 pt-1 text-xs">
          <button
            disabled={busy}
            onClick={() =>
              run(async () => {
                const key = await regenerateStreamKey(room.id);
                setStreamKey(key);
                setShowKey(true);
              })
            }
            className="rounded-lg bg-neutral-800 px-3 py-1.5 font-medium hover:bg-neutral-700 disabled:opacity-50"
          >
            Regenerate key
          </button>
          <button
            disabled={busy}
            onClick={() => run(async () => setIngest(await requestIngest(room.id)))}
            className="rounded-lg bg-neutral-800 px-3 py-1.5 font-medium hover:bg-neutral-700 disabled:opacity-50"
          >
            Get ingest URL
          </button>
          {room.status === "live" && (
            <button
              disabled={busy}
              onClick={() =>
                run(async () => {
                  await endStream(room.id);
                  router.refresh();
                })
              }
              className="rounded-lg bg-red-950 px-3 py-1.5 font-medium text-red-400 hover:bg-red-900"
            >
              End stream
            </button>
          )}
        </div>

        <div className="flex items-center gap-2 border-t border-white/10 pt-3 text-xs">
          <input
            value={banUserId}
            onChange={(e) => setBanUserId(e.target.value)}
            placeholder="User ID to ban/unban from chat"
            className="flex-1 rounded-lg border border-white/10 bg-neutral-950 px-2.5 py-1.5 outline-none focus:border-brand-500"
          />
          <button
            disabled={busy || !banUserId}
            onClick={() => run(() => banFromChat(room.id, banUserId, 24 * 60 * 60))}
            className="shrink-0 rounded-lg bg-neutral-800 px-3 py-1.5 font-medium hover:bg-neutral-700 disabled:opacity-50"
          >
            Ban 24h
          </button>
          <button
            disabled={busy || !banUserId}
            onClick={() => run(() => unbanFromChat(room.id, banUserId))}
            className="shrink-0 rounded-lg bg-neutral-800 px-3 py-1.5 font-medium hover:bg-neutral-700 disabled:opacity-50"
          >
            Unban
          </button>
        </div>
      </div>
    </div>
  );
}
