import { useLayoutEffect, useRef } from 'react';

const time = (iso) => new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
const day = (iso) => new Date(iso).toLocaleDateString();

export default function MessageList({ messages, me, hasMore, loadingMore, onLoadMore }) {
  const box = useRef(null);
  const stick = useRef(true); // follow new messages unless the user scrolled up
  const prev = useRef({ height: 0, firstId: null });

  useLayoutEffect(() => {
    const el = box.current;
    const firstId = messages[0]?.id ?? null;
    if (prev.current.firstId !== null && firstId !== prev.current.firstId && prev.current.height) {
      // Older messages were prepended: keep the reader where they were.
      el.scrollTop += el.scrollHeight - prev.current.height;
    } else if (stick.current) {
      el.scrollTop = el.scrollHeight;
    }
    prev.current = { height: el.scrollHeight, firstId };
  }, [messages]);

  function onScroll() {
    const el = box.current;
    stick.current = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
  }

  return (
    <div className="messages" ref={box} onScroll={onScroll}>
      {hasMore && (
        <button className="link" onClick={onLoadMore} disabled={loadingMore}>
          {loadingMore ? 'Loading…' : 'Load earlier messages'}
        </button>
      )}
      {messages.length === 0 && <p className="muted center-text">No messages yet. Say hello!</p>}
      {messages.map((m, i) => {
        const mine = m.username === me;
        const newDay = i === 0 || day(messages[i - 1].created_at) !== day(m.created_at);
        const sameAuthor = !newDay && messages[i - 1].username === m.username;
        return (
          <div key={m.id}>
            {newDay && <div className="day">{day(m.created_at)}</div>}
            <div className={`msg ${mine ? 'mine' : ''} ${sameAuthor ? 'cont' : ''}`}>
              {!mine && !sameAuthor && <div className="author">{m.username}</div>}
              <div className="bubble">
                <span className="text">{m.content}</span>
                <span className="time">{time(m.created_at)}</span>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
