import { useCallback, useEffect, useRef, useState } from "react";
import { listVideos } from "../api/videos";

const PAGE_SIZE = 24;

// Pages through /api/videos; refetches from page 1 when q/sort change.
export default function useInfiniteVideos({ q, sort }) {
  const [videos, setVideos] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const page = useRef(0);
  const inflight = useRef(null);

  const loadMore = useCallback(async () => {
    if (inflight.current) return;
    const ctrl = new AbortController();
    inflight.current = ctrl;
    setLoading(true);
    try {
      const data = await listVideos({ q, sort, page: page.current + 1, pageSize: PAGE_SIZE, signal: ctrl.signal });
      page.current += 1;
      setVideos((prev) => (page.current === 1 ? data.items : [...prev, ...data.items]));
      setTotal(data.total);
      setError(null);
    } catch (e) {
      if (e.name !== "AbortError") setError(e.message);
    } finally {
      if (inflight.current === ctrl) inflight.current = null;
      setLoading(false);
    }
  }, [q, sort]);

  useEffect(() => {
    inflight.current?.abort();
    inflight.current = null;
    page.current = 0;
    setVideos([]);
    setTotal(0);
    loadMore();
    return () => inflight.current?.abort();
  }, [loadMore]);

  return { videos, total, loading, error, hasMore: videos.length < total, loadMore };
}
