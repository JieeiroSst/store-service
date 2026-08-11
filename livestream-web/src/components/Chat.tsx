"use client";

import { useEffect, useRef, useState } from "react";
import type { ChatMessage, Session } from "@/lib/types";
import { colorForUsername } from "@/lib/color";

const WS_BASE_URL = process.env.NEXT_PUBLIC_LIVESTREAM_WS_URL ?? "ws://localhost:8080";

interface IncomingFrame extends Partial<ChatMessage> {
  error?: string;
}

export default function Chat({ roomId, session }: { roomId: string; session: Session | null }) {
  const [messages, setMessages] = useState<IncomingFrame[]>([]);
  const [draft, setDraft] = useState("");
  const wsRef = useRef<WebSocket | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const userId = session?.userId ?? `anon-${Math.random().toString(36).slice(2, 10)}`;
    const username = session?.username || "anonymous";
    const url = `${WS_BASE_URL}/ws/rooms/${roomId}/chat?userId=${encodeURIComponent(userId)}&username=${encodeURIComponent(username)}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const frame = JSON.parse(event.data) as IncomingFrame;
        setMessages((prev) => [...prev.slice(-199), frame]);
      } catch {
        // ignore malformed frames
      }
    };

    return () => ws.close();
  }, [roomId, session]);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [messages]);

  function send() {
    if (!draft.trim() || wsRef.current?.readyState !== WebSocket.OPEN) return;
    wsRef.current.send(JSON.stringify({ body: draft.trim() }));
    setDraft("");
  }

  return (
    <div className="flex h-[36rem] flex-col rounded-xl bg-neutral-900/60 ring-1 ring-white/10 lg:h-full">
      <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
        <h2 className="text-sm font-semibold">Stream chat</h2>
      </div>

      <div ref={scrollRef} className="scrollbar-thin flex-1 space-y-2 overflow-y-auto px-4 py-3 text-sm">
        {messages.length === 0 && (
          <p className="text-center text-xs text-neutral-600">
            Say hello - messages from everyone watching will show up here.
          </p>
        )}
        {messages.map((m, i) =>
          m.error ? (
            <p key={i} className="rounded bg-red-950/50 px-2 py-1 text-xs text-red-400">
              {m.error}
            </p>
          ) : (
            <p key={i} className="leading-relaxed">
              <span className="font-semibold" style={{ color: colorForUsername(m.username ?? "") }}>
                {m.username}
              </span>
              <span className="text-neutral-500">: </span>
              <span className="text-neutral-200">{m.body}</span>
            </p>
          ),
        )}
      </div>

      <div className="flex gap-2 border-t border-white/10 p-3">
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && send()}
          placeholder="Send a message"
          maxLength={500}
          className="flex-1 rounded-lg border border-white/10 bg-neutral-900 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />
        <button
          onClick={send}
          disabled={!draft.trim()}
          className="rounded-lg bg-brand-600 px-3 py-2 text-sm font-medium text-white transition hover:bg-brand-500 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Chat
        </button>
      </div>
    </div>
  );
}
