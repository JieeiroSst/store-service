"use client";

import { useEffect, useRef } from "react";
import Hls from "hls.js";
import { reportQoE } from "@/lib/apiClient";

const QOE_REPORT_INTERVAL_MS = 20_000;

export default function Player({ roomId, src }: { roomId: string; src: string }) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    let bufferingEvents = 0;
    let lastBitrateKbps = 0;

    const flush = () => {
      if (bufferingEvents > 0 || lastBitrateKbps > 0) {
        void reportQoE(roomId, lastBitrateKbps, bufferingEvents);
        bufferingEvents = 0;
      }
    };
    const interval = setInterval(flush, QOE_REPORT_INTERVAL_MS);

    let hls: Hls | null = null;
    if (Hls.isSupported()) {
      hls = new Hls();
      hls.loadSource(src);
      hls.attachMedia(video);
      hls.on(Hls.Events.LEVEL_SWITCHED, (_evt, data) => {
        const level = hls?.levels[data.level];
        if (level) lastBitrateKbps = Math.round(level.bitrate / 1000);
      });
      hls.on(Hls.Events.ERROR, (_evt, data) => {
        if (data.details === Hls.ErrorDetails.BUFFER_STALLED_ERROR) bufferingEvents += 1;
      });
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
      // Safari has native HLS support - no hls.js needed.
      video.src = src;
    }

    return () => {
      clearInterval(interval);
      flush();
      hls?.destroy();
    };
  }, [roomId, src]);

  return (
    <div className="overflow-hidden rounded-xl bg-black shadow-2xl ring-1 ring-white/10">
      <video ref={videoRef} controls autoPlay playsInline className="aspect-video w-full bg-black" />
    </div>
  );
}
