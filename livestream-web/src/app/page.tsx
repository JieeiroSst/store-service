import Link from "next/link";
import { getViewerCount, listRooms } from "@/lib/livestreamClient";
import RoomCard from "@/components/RoomCard";
import { gradientFor, initialsFor } from "@/lib/color";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const liveRooms = await listRooms(true);
  const withCounts = await Promise.all(
    liveRooms.map(async (room) => ({ room, viewerCount: await getViewerCount(room.id) })),
  );
  withCounts.sort((a, b) => b.viewerCount - a.viewerCount);
  const [featured, ...rest] = withCounts;

  if (!featured) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-24 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-fuchsia-500 shadow-glow">
          <span className="h-3 w-3 rounded-full bg-white" />
        </div>
        <h1 className="text-xl font-semibold">Nobody is live right now</h1>
        <p className="max-w-sm text-sm text-neutral-500">
          Check back soon, or head to your dashboard to start your own stream.
        </p>
        <Link
          href="/dashboard"
          className="mt-2 rounded-full bg-gradient-to-r from-brand-600 to-fuchsia-600 px-5 py-2 text-sm font-semibold shadow-glow"
        >
          Go to dashboard
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-3 flex items-center gap-2">
        <span className="h-2 w-2 animate-pulse rounded-full bg-red-500" />
        <h1 className="text-sm font-semibold uppercase tracking-wide text-neutral-400">Live now</h1>
      </div>

      {/* Featured stream */}
      <Link href={`/rooms/${featured.room.id}`} className="group mb-10 block">
        <div
          className="relative aspect-[16/7] overflow-hidden rounded-2xl transition group-hover:ring-2 group-hover:ring-brand-500"
          style={{ backgroundImage: gradientFor(featured.room.id) }}
        >
          <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent" />
          <span className="absolute left-4 top-4 flex items-center gap-1 rounded bg-red-600 px-2 py-1 text-xs font-bold uppercase tracking-wide text-white">
            <span className="h-1.5 w-1.5 rounded-full bg-white" />
            Live
          </span>
          <span className="absolute right-4 top-4 rounded bg-black/60 px-2 py-1 text-xs font-medium text-white backdrop-blur-sm">
            {featured.viewerCount.toLocaleString()} watching
          </span>
          <div className="absolute inset-x-0 bottom-0 flex items-end gap-3 p-5">
            <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-neutral-900/80 text-sm font-bold text-white ring-2 ring-white/20">
              {initialsFor(featured.room.title)}
            </span>
            <div className="min-w-0">
              <h2 className="truncate text-xl font-bold text-white">{featured.room.title}</h2>
              {featured.room.description && (
                <p className="truncate text-sm text-neutral-300">{featured.room.description}</p>
              )}
            </div>
          </div>
        </div>
      </Link>

      {rest.length > 0 && (
        <>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-neutral-500">
            More live channels
          </h2>
          <div className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2 lg:grid-cols-3">
            {rest.map(({ room, viewerCount }) => (
              <RoomCard key={room.id} room={room} viewerCount={viewerCount} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
