"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createRoom } from "@/lib/apiClient";

export default function CreateRoomForm() {
  const router = useRouter();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await createRoom(title, description);
      setTitle("");
      setDescription("");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to create room");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="mb-8 rounded-xl bg-neutral-900/60 p-5 ring-1 ring-white/10"
    >
      <h2 className="mb-3 text-sm font-semibold">Create a room</h2>
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Room title"
          required
          className="flex-1 rounded-lg border border-white/10 bg-neutral-950 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />
        <input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Description (optional)"
          className="flex-1 rounded-lg border border-white/10 bg-neutral-950 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />
        <button
          type="submit"
          disabled={submitting}
          className="rounded-lg bg-gradient-to-r from-brand-600 to-fuchsia-600 px-4 py-2 text-sm font-semibold text-white shadow-glow transition hover:opacity-90 disabled:opacity-50"
        >
          {submitting ? "Creating..." : "Create"}
        </button>
      </div>
      {error && <p className="mt-3 text-sm text-red-400">{error}</p>}
    </form>
  );
}
