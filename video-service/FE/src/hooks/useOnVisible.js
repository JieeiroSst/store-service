import { useEffect, useRef } from "react";

// Returns a ref; calls cb whenever the element scrolls into view (used for
// infinite scroll).
export default function useOnVisible(cb, enabled = true) {
  const ref = useRef(null);
  const cbRef = useRef(cb);
  cbRef.current = cb;

  useEffect(() => {
    const el = ref.current;
    if (!el || !enabled) return;
    const io = new IntersectionObserver(
      (entries) => entries[0].isIntersecting && cbRef.current(),
      { rootMargin: "400px" }
    );
    io.observe(el);
    return () => io.disconnect();
  }, [enabled]);

  return ref;
}
