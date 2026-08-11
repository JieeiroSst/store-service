import { redirect } from "next/navigation";
import { listRooms } from "@/lib/livestreamClient";
import { getSession } from "@/lib/session";
import CreateRoomForm from "@/components/CreateRoomForm";
import DashboardRoomCard from "@/components/DashboardRoomCard";

export const dynamic = "force-dynamic";

export default async function DashboardPage() {
  const session = getSession();
  if (!session) redirect("/login");

  const allRooms = await listRooms();
  const myRooms = allRooms.filter((r) => r.ownerUserId === session.userId);

  return (
    <div>
      <h1 className="mb-1 text-2xl font-bold">Your rooms</h1>
      <p className="mb-6 text-sm text-neutral-500">Manage stream keys, go live, and moderate chat.</p>
      <CreateRoomForm />
      {myRooms.length === 0 ? (
        <p className="text-sm text-neutral-500">You don&apos;t have any rooms yet - create one above.</p>
      ) : (
        <div className="flex flex-col gap-4">
          {myRooms.map((room) => (
            <DashboardRoomCard key={room.id} room={room} />
          ))}
        </div>
      )}
    </div>
  );
}
