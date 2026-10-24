import { Link, useLocation } from "react-router-dom";
import { Icon } from "./Icons";
import "./Sidebar.css";

const ITEMS = [
  { to: "/", label: "Home", icon: "home", isActive: (loc, sort) => loc.pathname === "/" && sort !== "popular" },
  { to: "/?sort=popular", label: "Trending", icon: "fire", isActive: (loc, sort) => loc.pathname === "/" && sort === "popular" },
];

// `overlay` = drawer over the content (watch page, phones); otherwise it takes
// up space beside it.
export default function Sidebar({ open, overlay, onClose }) {
  const loc = useLocation();
  const sort = new URLSearchParams(loc.search).get("sort");
  if (!open) return null;

  return (
    <>
      {overlay && <div className="scrim" onClick={onClose} />}
      <nav className={overlay ? "sidebar overlay" : "sidebar"}>
        {ITEMS.map((it) => (
          <Link
            key={it.label}
            to={it.to}
            className={it.isActive(loc, sort) ? "side-item active" : "side-item"}
            onClick={overlay ? onClose : undefined}
          >
            <Icon name={it.icon} />
            <span>{it.label}</span>
          </Link>
        ))}
      </nav>
    </>
  );
}
