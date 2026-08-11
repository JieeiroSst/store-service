import { redirect } from "next/navigation";
import { listRooms } from "@/lib/livestreamClient";
import { getSession } from "@/lib/session";
import AdminRoomTable from "@/components/AdminRoomTable";

export const dynamic = "force-dynamic";

export default async function AdminPage() {
  const session = getSession();
  if (!session || session.role !== "admin") redirect("/dashboard");

  const rooms = await listRooms();

  return (
    <div>
      <h1 className="mb-1 text-2xl font-bold">Admin</h1>
      <p className="mb-6 text-sm text-neutral-500">
        A room must be offline before it can be deleted - see livestream_service&apos;s
        ModerationUsecase.DeleteRoom.
      </p>
      <AdminRoomTable rooms={rooms} />
    </div>
  );
}
