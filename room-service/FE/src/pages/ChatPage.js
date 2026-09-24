import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate, useOutletContext, useParams } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import * as api from '../api/client';
import { useAuth } from '../auth/AuthContext';
import { useRoomSocket } from '../hooks/useRoomSocket';
import MemberPanel from '../components/MemberPanel';
import MessageInput from '../components/MessageInput';
import MessageList from '../components/MessageList';

const PAGE = 50;

// Union by id, oldest first: live events and history fetches can overlap.
function merge(existing, incoming) {
  const byId = new Map(existing.map((m) => [m.id, m]));
  incoming.forEach((m) => byId.set(m.id, m));
  return [...byId.values()].sort((a, b) => a.id - b.id);
}

export default function ChatPage() {
  const { roomId } = useParams();
  // Remount per room so no state (or in-flight response) leaks between rooms.
  return <Chat key={roomId} roomId={roomId} />;
}

function Chat({ roomId }) {
  const { user } = useAuth();
  const { reloadRooms } = useOutletContext();
  const navigate = useNavigate();

  const [room, setRoom] = useState(null);
  const [members, setMembers] = useState([]);
  const [messages, setMessages] = useState([]);
  const [hasMore, setHasMore] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [showMembers, setShowMembers] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const refreshMembers = useCallback(async () => {
    setMembers(await api.listMembers(roomId));
  }, [roomId]);

  useEffect(() => {
    let cancelled = false;
    Promise.all([api.getRoom(roomId), api.listMembers(roomId), api.listMessages(roomId, { limit: PAGE })])
      .then(([r, ms, msgs]) => {
        if (cancelled) return;
        setRoom(r);
        setMembers(ms);
        setMessages(msgs);
        setHasMore(msgs.length >= PAGE);
      })
      .catch((err) => {
        if (!cancelled) setError(err.status === 403 || err.status === 404 ? "This room doesn't exist or you're not a member." : err.message);
      });
    return () => {
      cancelled = true;
    };
  }, [roomId]);

  const onEvent = useCallback(
    (e) => {
      switch (e.type) {
        case 'message':
          setMessages((prev) => merge(prev, [e.message]));
          break;
        case 'member_added':
          refreshMembers().catch(() => {});
          reloadRooms();
          break;
        case 'member_removed':
          reloadRooms();
          if (e.member.username === user.username) navigate('/', { replace: true });
          else refreshMembers().catch(() => {});
          break;
        case 'error':
          setNotice(e.error);
          break;
        default:
      }
    },
    [refreshMembers, reloadRooms, navigate, user.username]
  );

  // After a reconnect, catch up on what was sent while disconnected.
  const onOpen = useCallback(
    (isReconnect) => {
      if (!isReconnect) return;
      api.listMessages(roomId, { limit: PAGE }).then((msgs) => setMessages((prev) => merge(prev, msgs))).catch(() => {});
      refreshMembers().catch(() => {});
    },
    [roomId, refreshMembers]
  );

  const { status, send } = useRoomSocket(roomId, { onEvent, onOpen });

  useEffect(() => {
    if (!notice) return undefined;
    const t = setTimeout(() => setNotice(''), 4000);
    return () => clearTimeout(t);
  }, [notice]);

  async function loadMore() {
    if (!messages.length) return;
    setLoadingMore(true);
    try {
      const older = await api.listMessages(roomId, { before: messages[0].id, limit: PAGE });
      setMessages((prev) => merge(prev, older));
      setHasMore(older.length >= PAGE);
    } catch (err) {
      setNotice(err.message);
    } finally {
      setLoadingMore(false);
    }
  }

  async function addMember(username) {
    await api.addMember(roomId, username);
    // The member_added event refreshes the list too; do it here so the
    // result is visible even if the socket is down.
    await refreshMembers();
    reloadRooms();
  }

  async function removeMember(username) {
    await api.removeMember(roomId, username);
    if (username === user.username) {
      reloadRooms();
      navigate('/', { replace: true });
    } else {
      await refreshMembers();
      reloadRooms();
    }
  }

  if (error) {
    return (
      <div className="center">
        <div className="card">
          <p>{error}</p>
          <Link to="/">Back to rooms</Link>
        </div>
      </div>
    );
  }
  if (!room) return <div className="center muted">Loading…</div>;

  const isOwner = room.owner === user.username;
  const connected = status === 'open';

  return (
    <div className="chat">
      <div className="chat-main">
        <header className="chat-header">
          <Link to="/" className="icon back" aria-label="Back to rooms">
            <ArrowLeft size={18} />
          </Link>
          <div className="grow">
            <strong>{room.name}</strong>
            <div className="small muted">
              <span className={`dot ${connected ? 'on' : 'off'}`} />
              {connected ? 'Connected' : status === 'connecting' ? 'Connecting…' : 'Reconnecting…'}
            </div>
          </div>
          <button className="link toggle-members" onClick={() => setShowMembers((s) => !s)}>
            {members.length} member{members.length === 1 ? '' : 's'}
          </button>
        </header>
        {notice && <div className="error banner" role="alert">{notice}</div>}
        <MessageList messages={messages} me={user.username} hasMore={hasMore} loadingMore={loadingMore} onLoadMore={loadMore} />
        <MessageInput onSend={send} disabled={!connected} />
      </div>
      <div className={`members-wrap ${showMembers ? 'open' : ''}`}>
        <MemberPanel members={members} me={user.username} isOwner={isOwner} onAdd={addMember} onRemove={removeMember} />
      </div>
    </div>
  );
}
