import { useEffect, useState } from "react";
import { Route, Routes, useLocation, useNavigate } from "react-router-dom";
import Header from "./components/Header";
import Sidebar from "./components/Sidebar";
import UploadModal from "./components/UploadModal";
import useTheme from "./hooks/useTheme";
import HomePage from "./pages/HomePage";
import WatchPage from "./pages/WatchPage";

const useIsNarrow = () => {
  const query = "(max-width: 900px)";
  const [narrow, setNarrow] = useState(() => window.matchMedia(query).matches);
  useEffect(() => {
    const m = window.matchMedia(query);
    const on = () => setNarrow(m.matches);
    m.addEventListener("change", on);
    return () => m.removeEventListener("change", on);
  }, []);
  return narrow;
};

export default function App() {
  const [theme, toggleTheme] = useTheme();
  const [menuOpen, setMenuOpen] = useState(true);
  const [uploading, setUploading] = useState(false);
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const narrow = useIsNarrow();

  // Like YouTube: the watch page and phones get the sidebar as a drawer, the
  // home page gets it docked. Every navigation closes the drawer.
  const onWatch = pathname.startsWith("/watch/");
  const overlay = onWatch || narrow;
  useEffect(() => setMenuOpen(!(onWatch || narrow)), [onWatch, narrow]);

  return (
    <>
      <Header
        onMenu={() => setMenuOpen((o) => !o)}
        onUpload={() => setUploading(true)}
        theme={theme}
        onToggleTheme={toggleTheme}
      />
      <div className="layout">
        <Sidebar open={menuOpen} overlay={overlay} onClose={() => setMenuOpen(false)} />
        <main className="content">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/watch/:id" element={<WatchPage />} />
            <Route path="*" element={<div className="empty">Page not found.</div>} />
          </Routes>
        </main>
      </div>

      {uploading && (
        <UploadModal
          onClose={() => setUploading(false)}
          onUploaded={(video) => {
            setUploading(false);
            navigate(`/watch/${video.id}`);
          }}
        />
      )}
    </>
  );
}
