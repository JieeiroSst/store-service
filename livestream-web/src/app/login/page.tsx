"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error ?? "login failed");
      }
      router.push(searchParams.get("next") ?? "/dashboard");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "login failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex min-h-[70vh] items-center justify-center">
      <div className="w-full max-w-sm rounded-2xl bg-neutral-900/60 p-8 shadow-2xl ring-1 ring-white/10">
        <div className="mb-6 flex flex-col items-center text-center">
          <span className="mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-fuchsia-500 shadow-glow">
            <span className="h-3 w-3 rounded-full bg-white" />
          </span>
          <h1 className="text-xl font-bold">Welcome back</h1>
          <p className="mt-1 text-sm text-neutral-500">Log in to go live or manage your rooms</p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-400">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              className="w-full rounded-lg border border-white/10 bg-neutral-950 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-neutral-400">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full rounded-lg border border-white/10 bg-neutral-950 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
          {error && (
            <p className="rounded-lg bg-red-950/50 px-3 py-2 text-sm text-red-400">{error}</p>
          )}
          <button
            type="submit"
            disabled={submitting}
            className="mt-2 rounded-lg bg-gradient-to-r from-brand-600 to-fuchsia-600 px-3 py-2.5 text-sm font-semibold text-white shadow-glow transition hover:opacity-90 disabled:opacity-50"
          >
            {submitting ? "Logging in..." : "Log in"}
          </button>
        </form>
        <p className="mt-5 text-center text-xs text-neutral-600">
          Authenticated against user_service - this app never issues its own accounts.
        </p>
      </div>
    </div>
  );
}
