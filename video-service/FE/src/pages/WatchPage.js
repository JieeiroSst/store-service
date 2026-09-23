import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { deleteVideo, getVideo, relatedVideos } from "../api/videos";
import { Icon } from "../components/Icons";
import Player from "../components/Player";
import { CardSkeleton, CompactCard } from "../components/VideoCard";
import { formatViews, timeAgo } from "../utils/format";
import "./WatchPage.css";

const POLL_MS = 4000;

export default function WatchPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [video, setVideo] = useState(null); // latest metadata
  const [playing, setPlaying] = useState(null); // what the player was started with
  const [related, setRelated] = useState(null);
  const [notFound, setNotFound] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const ctrl = new AbortController();
    setVideo(null);
    setPlaying(null);
    setRelated(null);
    setNotFound(false);
    setExpanded(false);
    window.scrollTo(0, 0);

    getVideo(id, ctrl.signal)
      .then((v) => {
        setVideo(v);
        setPlaying(v);
      })
      .catch((e) => e.status === 404 && setNotFound(true));
    relatedVideos(id, ctrl.signal)
      .then((r) => setRelated(r.items))
      .catch(() => setRelated([]));
    return () => ctrl.abort();
  }, [id]);

  // While transcoding, poll until the adaptive version is ready. The player
  // keeps playing the original meanwhile; the user switches when they like.
  useEffect(() => {
    if (!video || video.status !== "processing") return;
    const t = setInterval(() => getVideo(id).then(setVideo).catch(() => {}), POLL_MS);
    return () => clearInterval(t);
  }, [id, video]);

  useEffect(() => {
    document.title = video ? `${video.title || "Untitled"} - StreamTube` : "StreamTube";
  }, [video]);

  const remove = useCallback(async () => {
    if (!window.confirm(`Delete “${video.title || "Untitled"}”? This can't be undone.`)) return;
    await deleteVideo(id);
    navigate("/", { replace: true });
  }, [id, navigate, video]);

  const share = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (_) {}
  };

  if (notFound) return <div className="empty">This video isn't available.</div>;

  const upgradeReady = video?.status === "ready" && playing?.status !== "ready";

  return (
    <div className="watch">
      <div className="primary">
        {playing ? <Player video={playing} /> : <div className="player-skeleton skeleton" />}

        {playing && playing.status === "processing" && (
          <div className="banner">
            {upgradeReady ? (
              <>
                HD version is ready.
                <button className="btn primary" onClick={() => setPlaying(video)}>
                  Switch now
                </button>
              </>
            ) : (
              <>
                <span className="spinner" /> Processing higher qualities… you can watch the original meanwhile.
              </>
            )}
          </div>
        )}

        {video ? (
          <>
            <h1 className="title">{video.title || "Untitled"}</h1>
            <div className="actions">
              <span className="stats">
                {formatViews(video.views)} · {timeAgo(video.created_at)}
              </span>
              <div className="buttons">
                <button className="btn" onClick={share}>
                  <Icon name={copied ? "check" : "share"} size={20} />
                  {copied ? "Link copied" : "Share"}
                </button>
                <button className="btn danger" onClick={remove}>
                  <Icon name="trash" size={20} />
                  Delete
                </button>
              </div>
            </div>

            <div className={expanded ? "description expanded" : "description"} onClick={() => setExpanded(true)}>
              <strong>
                {formatViews(video.views)} · {new Date(video.created_at).toLocaleDateString(undefined, { dateStyle: "medium" })}
              </strong>
              <p>{video.description || "No description."}</p>
              {!expanded && video.description?.length > 140 && <span className="more">…more</span>}
            </div>
          </>
        ) : (
          <div className="skeleton title-skeleton" />
        )}
      </div>

      <aside className="secondary">
        {related === null && Array.from({ length: 6 }, (_, i) => <CardSkeleton key={i} compact />)}
        {related?.map((v) => (
          <CompactCard key={v.id} video={v} />
        ))}
        {related?.length === 0 && <p className="empty">No other videos yet.</p>}
      </aside>
    </div>
  );
}
