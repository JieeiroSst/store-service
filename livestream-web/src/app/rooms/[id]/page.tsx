import { notFound } from "next/navigation";
import { getPlaybackInfo, getRoom, getViewerCount } from "@/lib/livestreamClient";
import { getSession } from "@/lib/session";
import { initialsFor } from "@/lib/color";
import Player from "@/components/Player";
import Chat from "@/components/Chat";
import ViewerCount from "@/components/ViewerCount";

export const dynamic = "force-dynamic";

export default async function RoomPage({ params }: { params: { id: string } }) {
  const room = await getRoom(params.id);
  if (!room) notFound();

  const [playback, viewerCount, session] = await Promise.all([
    getPlaybackInfo(params.id),
    getViewerCount(params.id),
    Promise.resolve(getSession()),
  ]);

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-3 lg:items-start">
      <div className="lg:col-span-2">
        {playback ? (
          <Player roomId={room.id} src={playback.url} />
        ) : (
          <div className="flex aspect-video items-center justify-center rounded-xl bg-neutral-900 text-sm text-neutral-500 ring-1 ring-white/10">
            This room has nothing to play right now.
          </div>
        )}

        <div className="mt-4 flex items-start gap-3">
          <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-neutral-800 text-sm font-semibold text-neutral-300">
            {initialsFor(room.title)}
          </span>
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-lg font-bold">{room.title}</h1>
              {room.status === "live" ? (
                <span className="rounded bg-red-600 px-1.5 py-0.5 text-[11px] font-bold uppercase tracking-wide text-white">
                  Live
                </span>
              ) : (
                <span className="rounded bg-neutral-800 px-1.5 py-0.5 text-[11px] uppercase tracking-wide text-neutral-500">
                  Offline
                </span>
              )}
              <ViewerCount roomId={room.id} initialCount={viewerCount} />
            </div>
            {room.description && (
              <p className="mt-1 text-sm text-neutral-400">{room.description}</p>
            )}
          </div>
        </div>
      </div>

      <div className="lg:sticky lg:top-20 lg:h-[calc(100vh-6rem)]">
        <Chat roomId={room.id} session={session} />
      </div>
    </div>
  );
}
