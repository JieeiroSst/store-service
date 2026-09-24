import { useEffect, useRef, useState } from 'react';
import { accessToken, socketUrl } from '../api/client';

const MIN_DELAY = 1000;
const MAX_DELAY = 15000;

// Keeps a WebSocket open to one room, reconnecting with backoff. `onEvent`
// receives each server event; `onOpen(isReconnect)` fires after every connect
// so the caller can refetch whatever it missed while disconnected.
export function useRoomSocket(roomId, { onEvent, onOpen }) {
  const [status, setStatus] = useState('connecting');
  const socketRef = useRef(null);
  // Latest callbacks without re-opening the socket when they change.
  const handlers = useRef({ onEvent, onOpen });
  handlers.current = { onEvent, onOpen };

  useEffect(() => {
    let stopped = false;
    let attempt = 0;
    let timer = null;
    let opened = false;

    async function connect() {
      setStatus(opened ? 'reconnecting' : 'connecting');
      let token;
      try {
        token = await accessToken(); // refreshes if the last one expired
      } catch (err) {
        if (err.status === 401) return; // signed out; the app redirects to login
        return schedule();
      }
      if (stopped) return;

      const ws = new WebSocket(socketUrl(roomId, token));
      socketRef.current = ws;
      ws.onopen = () => {
        attempt = 0;
        setStatus('open');
        handlers.current.onOpen?.(opened);
        opened = true;
      };
      ws.onmessage = (e) => {
        try {
          handlers.current.onEvent?.(JSON.parse(e.data));
        } catch {
          // ignore malformed frames
        }
      };
      ws.onclose = () => {
        if (socketRef.current === ws) socketRef.current = null;
        if (!stopped) schedule();
      };
    }

    function schedule() {
      setStatus('reconnecting');
      const delay = Math.min(MAX_DELAY, MIN_DELAY * 2 ** attempt++);
      timer = setTimeout(connect, delay);
    }

    connect();
    return () => {
      stopped = true;
      clearTimeout(timer);
      socketRef.current?.close();
      socketRef.current = null;
    };
  }, [roomId]);

  const send = (content) => {
    const ws = socketRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return false;
    ws.send(JSON.stringify({ content }));
    return true;
  };

  return { status, send };
}
