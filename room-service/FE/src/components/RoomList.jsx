import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus } from 'lucide-react';

export default function RoomList({ rooms, activeId, onCreate }) {
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  async function submit(e) {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      await onCreate(name.trim());
      setName('');
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <form className="row create-room" onSubmit={submit}>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="New room name" maxLength={100} aria-label="New room name" />
        <button className="icon primary" disabled={busy || !name.trim()} title="Create room" aria-label="Create room">
          <Plus size={18} />
        </button>
      </form>
      {error && <div className="error">{error}</div>}
      <nav className="room-list">
        {rooms.length === 0 && <p className="muted pad">No rooms yet.</p>}
        {rooms.map((r) => (
          <Link key={r.id} to={`/rooms/${r.id}`} className={String(r.id) === activeId ? 'active' : ''}>
            <span className="room-name">{r.name}</span>
            <span className="muted small">{r.member_count} member{r.member_count === 1 ? '' : 's'}</span>
          </Link>
        ))}
      </nav>
    </>
  );
}
