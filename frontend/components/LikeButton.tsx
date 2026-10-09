"use client";

import { useEffect, useState } from "react";
import { api, refreshResources } from "../utils/api";

export default function LikeButton({
  id,
  count,
  liked,
}: {
  id: number;
  count: number;
  liked: boolean;
}) {
  const [value, setValue] = useState({ count, liked });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    setValue({ count, liked });
  }, [count, liked]);
  const toggle = async () => {
    setBusy(true);
    setError("");
    try {
      const result = await api<{ likeCount: number; isLiked: boolean }>(
        `/posts/like?id=${id}`,
        { method: value.liked ? "DELETE" : "PUT" },
      );
      setValue({ count: result.likeCount, liked: result.isLiked });
      refreshResources("posts");
    } catch (error) {
      setError(
        error instanceof Error ? error.message : "Unable to update like.",
      );
    } finally {
      setBusy(false);
    }
  };
  return (
    <>
      <button
        type="button"
        className={`interaction-btn ${value.liked ? "active" : ""}`}
        aria-pressed={value.liked}
        aria-label={`${value.liked ? "Unlike" : "Like"} post`}
        onClick={toggle}
        disabled={busy}
      >
        ♥ {value.count}
      </button>
      {error && <span role="alert">{error}</span>}
    </>
  );
}
