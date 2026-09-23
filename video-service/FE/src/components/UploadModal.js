import { useRef, useState } from "react";
import { uploadVideo } from "../api/videos";
import { formatSize } from "../utils/format";
import { Icon } from "./Icons";
import "./UploadModal.css";

export default function UploadModal({ onClose, onUploaded }) {
  const [file, setFile] = useState(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [progress, setProgress] = useState(null); // null = not started
  const [error, setError] = useState(null);
  const [dragging, setDragging] = useState(false);
  const input = useRef(null);
  const uploading = progress !== null;

  const choose = (f) => {
    if (!f) return;
    if (!f.type.startsWith("video/")) {
      setError("Please choose a video file.");
      return;
    }
    setError(null);
    setFile(f);
    setTitle((t) => t || f.name.replace(/\.[^.]+$/, ""));
  };

  const submit = async (e) => {
    e.preventDefault();
    if (!file) return;
    setError(null);
    setProgress(0);
    try {
      const video = await uploadVideo({ title: title.trim(), description, file }, setProgress);
      onUploaded(video);
    } catch (err) {
      setProgress(null);
      setError(err.message);
    }
  };

  return (
    <div className="modal-scrim" onMouseDown={(e) => e.target === e.currentTarget && !uploading && onClose()}>
      <form className="modal" onSubmit={submit}>
        <div className="modal-head">
          <h2>Upload video</h2>
          <button type="button" className="icon-btn" onClick={onClose} disabled={uploading} aria-label="Close">
            <Icon name="close" />
          </button>
        </div>

        {!file ? (
          <div
            className={dragging ? "drop dragging" : "drop"}
            onDragOver={(e) => {
              e.preventDefault();
              setDragging(true);
            }}
            onDragLeave={() => setDragging(false)}
            onDrop={(e) => {
              e.preventDefault();
              setDragging(false);
              choose(e.dataTransfer.files[0]);
            }}
          >
            <Icon name="upload" size={48} />
            <p>Drag and drop a video file here</p>
            <button type="button" className="btn primary" onClick={() => input.current.click()}>
              Select file
            </button>
            <input ref={input} type="file" accept="video/*" hidden onChange={(e) => choose(e.target.files[0])} />
          </div>
        ) : (
          <div className="fields">
            <p className="file-info">
              {file.name} · {formatSize(file.size)}
            </p>
            <label>
              Title
              <input value={title} maxLength={200} onChange={(e) => setTitle(e.target.value)} disabled={uploading} />
            </label>
            <label>
              Description
              <textarea
                rows={5}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={uploading}
                placeholder="Tell viewers about your video"
              />
            </label>
            {uploading && (
              <div className="progress" role="progressbar" aria-valuenow={Math.round(progress * 100)}>
                <div style={{ width: `${progress * 100}%` }} />
                <span>{progress < 1 ? `Uploading ${Math.round(progress * 100)}%` : "Saving…"}</span>
              </div>
            )}
          </div>
        )}

        {error && <p className="modal-error">{error}</p>}

        {file && (
          <div className="modal-foot">
            <button type="button" className="btn" onClick={() => setFile(null)} disabled={uploading}>
              Change file
            </button>
            <button type="submit" className="btn primary" disabled={uploading || !title.trim()}>
              {uploading ? "Uploading…" : "Upload"}
            </button>
          </div>
        )}
      </form>
    </div>
  );
}
