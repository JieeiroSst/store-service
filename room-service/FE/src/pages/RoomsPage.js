import { useCallback, useEffect, useState } from 'react';
import { Outlet, useNavigate, useParams } from 'react-router-dom';
import { LogOut } from 'lucide-react';
import * as api from '../api/client';
import { useAuth } from '../auth/AuthContext';
import RoomList from '../components/RoomList';

// Layout: the caller's rooms on the left, the open room (Outlet) on the right.
export default function RoomsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const { roomId } = useParams();
  const [rooms, setRooms] = useState([]);
  const [error, setError] = useState('');

  const reload = useCallback(async () => {
    try {
      setRooms(await api.listRooms());
      setError('');
    } catch (err) {
      setError(err.message);
    }
  }, []);

  useEffect(() => {
    reload();
  }, [reload]);

  async function create(name) {
    const room = await api.createRoom(name);
    await reload();
    navigate(`/rooms/${room.id}`);
  }

  return (
    <div className={`layout ${roomId ? 'has-room' : ''}`}>
      <aside className="sidebar">
        <header>
          <strong title={user.username}>{user.username}</strong>
          <button className="icon" onClick={logout} title="Sign out" aria-label="Sign out">
            <LogOut size={18} />
          </button>
        </header>
        {error && <div className="error">{error}</div>}
        <RoomList rooms={rooms} activeId={roomId} onCreate={create} />
      </aside>
      <main className="main">
        <Outlet context={{ reloadRooms: reload }} />
      </main>
    </div>
  );
}
