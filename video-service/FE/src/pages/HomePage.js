import { useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { CardSkeleton, VideoCard } from "../components/VideoCard";
import useInfiniteVideos from "../hooks/useInfiniteVideos";
import useOnVisible from "../hooks/useOnVisible";
import "./HomePage.css";

const CHIPS = [
  { label: "All", sort: "newest" },
  { label: "Popular", sort: "popular" },
];

export default function HomePage() {
  const [params, setParams] = useSearchParams();
  const q = params.get("q") || "";
  const sort = params.get("sort") === "popular" ? "popular" : "newest";
  const { videos, loading, error, hasMore, loadMore } = useInfiniteVideos({ q, sort });
  const sentinel = useOnVisible(loadMore, hasMore && !loading);

  useEffect(() => {
    document.title = q ? `${q} - StreamTube` : "StreamTube";
  }, [q]);

  const setSort = (s) => {
    const next = new URLSearchParams(params);
    if (s === "popular") next.set("sort", s);
    else next.delete("sort");
    setParams(next);
  };

  return (
    <div className="home">
      <div className="chips">
        {CHIPS.map((c) => (
          <button key={c.sort} className={sort === c.sort ? "chip active" : "chip"} onClick={() => setSort(c.sort)}>
            {c.label}
          </button>
        ))}
        {q && <span className="results-for">Results for “{q}”</span>}
      </div>

      {error && <div className="error-box">{error}</div>}

      <div className="grid">
        {videos.map((v) => (
          <VideoCard key={v.id} video={v} />
        ))}
        {loading && Array.from({ length: videos.length ? 4 : 12 }, (_, i) => <CardSkeleton key={`s${i}`} />)}
      </div>

      {!loading && !error && videos.length === 0 && (
        <div className="empty">{q ? `No videos match “${q}”.` : "No videos yet. Click Create to upload the first one."}</div>
      )}

      <div ref={sentinel} style={{ height: 1 }} />
    </div>
  );
}
