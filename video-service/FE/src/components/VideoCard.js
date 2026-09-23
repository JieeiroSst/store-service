import { Link } from "react-router-dom";
import { urls } from "../api/videos";
import { formatDuration, formatViews, timeAgo } from "../utils/format";
import "./VideoCard.css";

function Thumb({ video }) {
  const ready = video.status === "ready";
  return (
    <div className="thumb">
      {ready ? (
        <img src={urls.thumbnail(video.id)} alt="" loading="lazy" />
      ) : (
        <div className="thumb-pending">
          <span className="spinner" />
          Processing…
        </div>
      )}
      {ready && video.duration > 0 && <span className="duration">{formatDuration(video.duration)}</span>}
    </div>
  );
}

// Home-page card: thumbnail on top.
export function VideoCard({ video }) {
  return (
    <Link to={`/watch/${video.id}`} className="card">
      <Thumb video={video} />
      <div className="card-meta">
        <h3 className="card-title">{video.title || "Untitled"}</h3>
        <p>
          {formatViews(video.views)} · {timeAgo(video.created_at)}
        </p>
      </div>
    </Link>
  );
}

// Watch-page sidebar row: thumbnail on the left.
export function CompactCard({ video }) {
  return (
    <Link to={`/watch/${video.id}`} className="card compact">
      <Thumb video={video} />
      <div className="card-meta">
        <h3 className="card-title">{video.title || "Untitled"}</h3>
        <p>
          {formatViews(video.views)} · {timeAgo(video.created_at)}
        </p>
      </div>
    </Link>
  );
}

export function CardSkeleton({ compact }) {
  return (
    <div className={compact ? "card compact" : "card"}>
      <div className="thumb skeleton" />
      <div className="card-meta">
        <div className="skeleton line" />
        <div className="skeleton line short" />
      </div>
    </div>
  );
}
