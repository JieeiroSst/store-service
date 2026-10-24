import { useState } from 'react';
import { UserPlus, X } from 'lucide-react';

// The owner adds people by username (user-service has no user search, so a
// typo just creates an invitation nobody can claim) and can remove anyone;
// everyone else can only leave.
export default function MemberPanel({ members, me, isOwner, onAdd, onRemove }) {
  const [username, setUsername] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  async function add(e) {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      await onAdd(username.trim());
      setUsername('');
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  async function remove(username) {
    setError('');
    try {
      await onRemove(username);
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <aside className="members">
      <h3>Members ({members.length})</h3>
      {isOwner && (
        <form className="row" onSubmit={add}>
          <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="Add by username" maxLength={64} aria-label="Username to add" />
          <button className="icon primary" disabled={busy || !username.trim()} title="Add member" aria-label="Add member">
            <UserPlus size={18} />
          </button>
        </form>
      )}
      {error && <div className="error">{error}</div>}
      <ul>
        {members.map((m) => (
          <li key={m.username}>
            <span>
              {m.username}
              {m.username === me && <span className="muted"> (you)</span>}
              {m.role === 'owner' && <span className="badge">owner</span>}
            </span>
            {m.role !== 'owner' && (isOwner || m.username === me) && (
              <button className="icon" onClick={() => remove(m.username)} title={m.username === me ? 'Leave room' : 'Remove'} aria-label={m.username === me ? 'Leave room' : `Remove ${m.username}`}>
                <X size={16} />
              </button>
            )}
          </li>
        ))}
      </ul>
    </aside>
  );
}
