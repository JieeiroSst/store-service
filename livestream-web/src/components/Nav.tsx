"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { Session } from "@/lib/types";
import { initialsFor } from "@/lib/color";

function NavLink({ href, active, children }: { href: string; active: boolean; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className={`rounded-full px-3 py-1.5 text-sm font-medium transition ${
        active ? "bg-white/10 text-white" : "text-neutral-400 hover:text-white"
      }`}
    >
      {children}
    </Link>
  );
}

export default function Nav({ session }: { session: Session | null }) {
  const router = useRouter();
  const pathname = usePathname();

  async function logout() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/");
    router.refresh();
  }

  return (
    <header className="sticky top-0 z-30 border-b border-white/5 bg-neutral-950/80 backdrop-blur-md">
      <nav className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3 sm:px-6">
        <Link href="/" className="flex items-center gap-2">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-fuchsia-500 shadow-glow">
            <span className="h-2.5 w-2.5 rounded-full bg-white" />
          </span>
          <span className="text-lg font-bold tracking-tight">
            live<span className="text-brand-400">stream</span>
          </span>
        </Link>

        <div className="flex items-center gap-1">
          <NavLink href="/" active={pathname === "/"}>
            Watch
          </NavLink>
          {session && (
            <NavLink href="/dashboard" active={pathname?.startsWith("/dashboard") ?? false}>
              Dashboard
            </NavLink>
          )}
          {session?.role === "admin" && (
            <NavLink href="/admin" active={pathname?.startsWith("/admin") ?? false}>
              Admin
            </NavLink>
          )}
        </div>

        <div className="flex items-center gap-3">
          {session ? (
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <span className="flex h-7 w-7 items-center justify-center rounded-full bg-neutral-800 text-xs font-semibold text-neutral-300">
                  {initialsFor(session.username || session.userId)}
                </span>
                <span className="hidden text-sm text-neutral-300 sm:inline">
                  {session.username || session.userId}
                </span>
              </div>
              <button
                onClick={logout}
                className="rounded-full border border-white/10 px-3 py-1.5 text-sm text-neutral-300 transition hover:border-white/20 hover:text-white"
              >
                Log out
              </button>
            </div>
          ) : (
            <Link
              href="/login"
              className="rounded-full bg-gradient-to-r from-brand-600 to-fuchsia-600 px-4 py-1.5 text-sm font-semibold text-white shadow-glow transition hover:opacity-90"
            >
              Log in
            </Link>
          )}
        </div>
      </nav>
    </header>
  );
}
