import Hls from "hls.js";
import { useCallback, useEffect, useRef, useState } from "react";
import { recordView, urls } from "../api/videos";
import { Icon } from "./Icons";
import "./Player.css";

const VIEW_AFTER_SECONDS = 5;
const RESUME_KEY = (id) => `resume:${id}`;

function readResume(id) {
  try {
    return parseFloat(localStorage.getItem(RESUME_KEY(id))) || 0;
  } catch (_) {
    return 0;
  }
}

/**
 * Plays a video the way YouTube does:
 *  - ready videos stream as HLS with adaptive quality (hls.js, or the browser's
 *    own HLS on Safari) and a quality menu;
 *  - videos still being transcoded play the original file (range requests)
 *    until the HLS version is ready.
 */
export default function Player({ video }) {
  const ref = useRef(null);
  const hlsRef = useRef(null);
  const [levels, setLevels] = useState([]); // [{index, height}]
  const [current, setCurrent] = useState(-1); // -1 = auto
  const [menu, setMenu] = useState(false);
  const [error, setError] = useState(null);
  const ready = video.status === "ready";

  // Attach the source.
  useEffect(() => {
    const el = ref.current;
    setLevels([]);
    setCurrent(-1);
    setError(null);
    let cleanup = () => {};

    const resume = () => {
      const t = readResume(video.id);
      // Don't resume in the last few seconds: that's "watched", start over.
      if (t > 1 && (!el.duration || t < el.duration - 5)) el.currentTime = t;
    };
    el.addEventListener("loadedmetadata", resume, { once: true });

    if (ready && Hls.isSupported()) {
      const hls = new Hls({ capLevelToPlayerSize: true });
      hlsRef.current = hls;
      hls.loadSource(urls.hls(video.id));
      hls.attachMedia(el);
      hls.on(Hls.Events.MANIFEST_PARSED, (_, data) => {
        setLevels(data.levels.map((l, index) => ({ index, height: l.height })).reverse());
        el.play().catch(() => {}); // autoplay may be blocked; controls still work
      });
      hls.on(Hls.Events.ERROR, (_, data) => {
        if (!data.fatal) return;
        if (data.type === Hls.ErrorTypes.NETWORK_ERROR) hls.startLoad();
        else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) hls.recoverMediaError();
        else setError("This video can't be played.");
      });
      cleanup = () => {
        hls.destroy();
        hlsRef.current = null;
      };
    } else if (ready && el.canPlayType("application/vnd.apple.mpegurl")) {
      el.src = urls.hls(video.id); // Safari/iOS play HLS natively
      el.play().catch(() => {});
    } else {
      el.src = urls.stream(video.id);
      el.play().catch(() => {});
    }

    return () => {
      cleanup();
      el.removeEventListener("loadedmetadata", resume);
      el.removeAttribute("src");
      el.load();
    };
  }, [video.id, ready]);

  // Count a view after a few seconds of actual playback; remember position.
  useEffect(() => {
    const el = ref.current;
    let counted = false;
    let lastSave = 0;
    const onTime = () => {
      if (!counted && el.currentTime >= VIEW_AFTER_SECONDS) {
        counted = true;
        recordView(video.id);
      }
      if (Math.abs(el.currentTime - lastSave) >= 5) {
        lastSave = el.currentTime;
        try {
          localStorage.setItem(RESUME_KEY(video.id), String(el.currentTime));
        } catch (_) {}
      }
    };
    el.addEventListener("timeupdate", onTime);
    return () => el.removeEventListener("timeupdate", onTime);
  }, [video.id]);

  // YouTube-style keyboard shortcuts.
  useEffect(() => {
    const onKey = (e) => {
      const el = ref.current;
      if (!el || e.metaKey || e.ctrlKey || e.altKey) return;
      if (/^(INPUT|TEXTAREA|SELECT)$/.test(e.target.tagName) || e.target.isContentEditable) return;
      const seek = (s) => (el.currentTime = Math.max(0, Math.min(el.duration || Infinity, el.currentTime + s)));
      switch (e.key) {
        case " ":
        case "k":
          el.paused ? el.play() : el.pause();
          break;
        case "j":
          seek(-10);
          break;
        case "l":
          seek(10);
          break;
        case "ArrowLeft":
          seek(-5);
          break;
        case "ArrowRight":
          seek(5);
          break;
        case "m":
          el.muted = !el.muted;
          break;
        case "f":
          document.fullscreenElement ? document.exitFullscreen() : el.parentElement.requestFullscreen();
          break;
        default:
          return;
      }
      e.preventDefault();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const pick = useCallback((index) => {
    if (hlsRef.current) hlsRef.current.currentLevel = index; // -1 = auto
    setCurrent(index);
    setMenu(false);
  }, []);

  const label = current === -1 ? "Auto" : `${levels.find((l) => l.index === current)?.height}p`;

  return (
    <div className="player">
      <video ref={ref} controls playsInline poster={ready ? urls.thumbnail(video.id) : undefined} />

      {levels.length > 1 && (
        <div className="quality">
          <button className="icon-btn" onClick={() => setMenu((m) => !m)} aria-label="Quality">
            <Icon name="gear" />
          </button>
          {menu && (
            <ul className="quality-menu">
              <li onClick={() => pick(-1)}>
                <span className="tick">{current === -1 && <Icon name="check" size={16} />}</span>Auto
              </li>
              {levels.map((l) => (
                <li key={l.index} onClick={() => pick(l.index)}>
                  <span className="tick">{current === l.index && <Icon name="check" size={16} />}</span>
                  {l.height}p
                </li>
              ))}
            </ul>
          )}
          <span className="quality-label">{label}</span>
        </div>
      )}

      {error && <div className="player-error">{error}</div>}
    </div>
  );
}
