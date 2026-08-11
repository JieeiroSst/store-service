"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { Room } from "@/lib/types";
import { adminDeleteRoom } from "@/lib/apiClient";
import { initialsFor } from "@/lib/color";

export default function AdminRoomTable({ rooms }: { rooms: Room[] }) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  async function handleDelete(id: string) {
    setError(null);
    setBusyId(id);
    try {
      await adminDeleteRoom(id);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to delete room");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="overflow-hidden rounded-xl bg-neutral-900/60 ring-1 ring-white/10">
      {error && (
        <p className="border-b border-white/10 bg-red-950/40 px-4 py-2 text-sm text-red-400">{error}</p>
      )}
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-white/10 text-xs uppercase tracking-wide text-neutral-500">
            <th className="px-4 py-3 font-medium">Room</th>
            <th className="px-4 py-3 font-medium">Owner</th>
            <th className="px-4 py-3 font-medium">Status</th>
            <th className="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody>
          {rooms.map((room) => (
            <tr key={room.id} className="border-b border-white/5 last:border-0 hover:bg-white/5">
              <td className="px-4 py-3">
                <div className="flex items-center gap-2.5">
                  <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-neutral-800 text-xs font-semibold text-neutral-300">
                    {initialsFor(room.title)}
                  </span>
                  <span className="truncate font-medium">{room.title}</span>
                </div>
              </td>
              <td className="px-4 py-3 text-neutral-400">{room.ownerUserId}</td>
              <td className="px-4 py-3">
                {room.status === "live" ? (
                  <span className="rounded bg-red-600 px-1.5 py-0.5 text-[11px] font-bold uppercase tracking-wide text-white">
                    Live
                  </span>
                ) : (
                  <span className="rounded bg-neutral-800 px-1.5 py-0.5 text-[11px] uppercase tracking-wide text-neutral-500">
                    Offline
                  </span>
                )}
              </td>
              <td className="px-4 py-3 text-right">
                <button
                  disabled={busyId === room.id}
                  onClick={() => handleDelete(room.id)}
                  className="rounded-lg bg-red-950 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-900 disabled:opacity-50"
                  title={room.status === "live" ? "End the stream before deleting" : undefined}
                >
                  Delete
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {rooms.length === 0 && (
        <p className="px-4 py-8 text-center text-sm text-neutral-500">No rooms yet.</p>
      )}
    </div>
  );
}
