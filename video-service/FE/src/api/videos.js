// Empty in dev (CRA proxies /api to the backend, see package.json "proxy") and
// in the container (nginx proxies it, and the ingress routes /api to the API).
const API = process.env.REACT_APP_API_URL || "";

async function json(res) {
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      message = (await res.json()).error || message;
    } catch (_) {}
    const err = new Error(message);
    err.status = res.status;
    throw err;
  }
  return res.status === 204 ? null : res.json();
}

export const urls = {
  thumbnail: (id) => `${API}/api/videos/${id}/thumbnail`,
  stream: (id) => `${API}/api/videos/${id}/stream`,
  hls: (id) => `${API}/api/videos/${id}/hls/master.m3u8`,
};

export const listVideos = ({ q = "", sort = "newest", page = 1, pageSize = 24, signal } = {}) => {
  const params = new URLSearchParams({ sort, page, page_size: pageSize });
  if (q) params.set("q", q);
  return fetch(`${API}/api/videos?${params}`, { signal }).then(json);
};

export const getVideo = (id, signal) => fetch(`${API}/api/videos/${id}`, { signal }).then(json);

export const relatedVideos = (id, signal) =>
  fetch(`${API}/api/videos/${id}/related?limit=15`, { signal }).then(json);

export const recordView = (id) => fetch(`${API}/api/videos/${id}/view`, { method: "POST" }).catch(() => {});

export const deleteVideo = (id) => fetch(`${API}/api/videos/${id}`, { method: "DELETE" }).then(json);

// XHR rather than fetch: only XHR reports upload progress.
export function uploadVideo({ title, description, file }, onProgress) {
  return new Promise((resolve, reject) => {
    // "title" and "description" must come before "file": the backend streams
    // the file straight to storage and reads fields in order.
    const form = new FormData();
    form.append("title", title);
    form.append("description", description);
    form.append("file", file);

    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${API}/api/videos`);
    xhr.upload.onprogress = (e) => e.lengthComputable && onProgress(e.loaded / e.total);
    xhr.onerror = () => reject(new Error("Network error"));
    xhr.onload = () => {
      let body = null;
      try {
        body = JSON.parse(xhr.responseText);
      } catch (_) {}
      if (xhr.status === 201) resolve(body);
      else reject(new Error(body?.error || `Upload failed (${xhr.status})`));
    };
    xhr.send(form);
  });
}
