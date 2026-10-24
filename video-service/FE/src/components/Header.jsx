import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { Icon } from "./Icons";
import "./Header.css";

export default function Header({ onMenu, onUpload, theme, onToggleTheme }) {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [q, setQ] = useState(params.get("q") || "");

  // Keep the box in sync with back/forward navigation.
  useEffect(() => setQ(params.get("q") || ""), [params]);

  const submit = (e) => {
    e.preventDefault();
    const term = q.trim();
    navigate(term ? `/?q=${encodeURIComponent(term)}` : "/");
  };

  return (
    <header className="header">
      <div className="header-left">
        <button className="icon-btn" onClick={onMenu} aria-label="Menu">
          <Icon name="menu" />
        </button>
        <Link to="/" className="logo" aria-label="Home">
          <span className="logo-mark">▶</span>
          <span className="logo-text">StreamTube</span>
        </Link>
      </div>

      <form className="search" onSubmit={submit} role="search">
        <input
          type="search"
          placeholder="Search"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search"
        />
        <button type="submit" aria-label="Search">
          <Icon name="search" />
        </button>
      </form>

      <div className="header-right">
        <button className="btn create" onClick={onUpload}>
          <Icon name="plus" />
          <span>Create</span>
        </button>
        <button className="icon-btn" onClick={onToggleTheme} aria-label="Toggle theme">
          <Icon name={theme === "dark" ? "sun" : "moon"} />
        </button>
      </div>
    </header>
  );
}
