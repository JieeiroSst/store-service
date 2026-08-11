import Link from "next/link";
import type { Room } from "@/lib/types";
import { gradientFor, initialsFor } from "@/lib/color";

export default function RoomCard({ room, viewerCount }: { room: Room; viewerCount?: number }) {
  return (
    <Link href={`/rooms/${room.id}`} className="group block">
      <div
        className="relative aspect-video overflow-hidden rounded-xl transition group-hover:ring-2 group-hover:ring-brand-500"
        style={{ backgroundImage: gradientFor(room.id) }}
      >
        <div className="absolute inset-0 bg-black/10 transition group-hover:bg-black/0" />
        {room.status === "live" && (
          <span className="absolute left-2 top-2 flex items-center gap-1 rounded bg-red-600 px-1.5 py-0.5 text-[11px] font-bold uppercase tracking-wide text-white">
            <span className="h-1.5 w-1.5 rounded-full bg-white" />
            Live
          </span>
        )}
        {room.status === "live" && viewerCount !== undefined && (
          <span className="absolute bottom-2 left-2 rounded bg-black/60 px-1.5 py-0.5 text-xs font-medium text-white backdrop-blur-sm">
            {viewerCount.toLocaleString()} watching
          </span>
        )}
      </div>
      <div className="mt-2 flex gap-2.5">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-neutral-800 text-xs font-semibold text-neutral-300">
          {initialsFor(room.title)}
        </span>
        <div className="min-w-0">
          <h3 className="truncate text-sm font-semibold text-neutral-100 group-hover:text-white">
            {room.title}
          </h3>
          {room.description && (
            <p className="truncate text-xs text-neutral-500">{room.description}</p>
          )}
        </div>
      </div>
    </Link>
  );
}
