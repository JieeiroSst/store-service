import { useState } from 'react';
import { Send } from 'lucide-react';

export default function MessageInput({ onSend, disabled }) {
  const [text, setText] = useState('');

  function submit(e) {
    e.preventDefault();
    const content = text.trim();
    if (content && onSend(content)) setText('');
  }

  // Enter sends, Shift+Enter adds a newline.
  function onKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing) submit(e);
  }

  return (
    <form className="composer" onSubmit={submit}>
      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder={disabled ? 'Reconnecting…' : 'Write a message'}
        rows={1}
        maxLength={4000}
        aria-label="Message"
      />
      <button className="icon primary" disabled={disabled || !text.trim()} title="Send" aria-label="Send">
        <Send size={18} />
      </button>
    </form>
  );
}
