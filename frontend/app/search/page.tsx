"use client";
import { FormEvent, useRef, useState } from "react";
import Link from "next/link";
import FeatureShell from "../../components/FeatureShell";
import ContentActions from "../../components/ContentActions";
import LikeButton from "../../components/LikeButton";
import { api } from "../../utils/api";
import { useRealtimeRefresh } from "../webscoket/useRealtimeRefresh";
type Item = {
  id: number;
  title?: string;
  description?: string;
  content?: string;
  userId?: number;
  likeCount?: number;
  isLiked?: boolean;
};
export default function Search() {
  const [term, setTerm] = useState("");
  const [kind, setKind] = useState("users");
  const [shown, setShown] = useState({ term: "", kind: "users" });
  const [items, setItems] = useState<Item[]>([]);
  const [more, setMore] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const query = useRef({ term: "", kind: "users", page: 1 });
  const sequence = useRef(0);
  async function run(page = 1, quiet = false) {
    const q = query.current;
    if (!q.term) return;
    const seq = ++sequence.current;
    if (!quiet) setBusy(true);
    setError("");
    try {
      const responses = await Promise.all(
        Array.from({ length: page }, (_, i) =>
          api<{ items: Item[]; hasMore?: boolean; total?: number }>(
            `/search?q=${encodeURIComponent(q.term)}&type=${q.kind}&page=${i + 1}`,
          ),
        ),
      );
      if (seq !== sequence.current) return;
      const all = responses.flatMap((r) => r.items);
      setShown({ term: q.term, kind: q.kind });
      setItems(Array.from(new Map(all.map((i) => [i.id, i])).values()));
      const last = responses[responses.length - 1];
      setMore(last.hasMore ?? all.length < (last.total || 0));
      query.current.page = page;
    } catch (e) {
      if (seq === sequence.current)
        setError(e instanceof Error ? e.message : "Search failed");
    } finally {
      if (seq === sequence.current) setBusy(false);
    }
  }
  useRealtimeRefresh(["users", "posts", "groups", "profiles", "social"], () =>
    run(query.current.page, true),
  );
  function submit(e: FormEvent) {
    e.preventDefault();
    query.current = { term: term.trim(), kind, page: 1 };
    run();
  }
  return (
    <FeatureShell title="Search" activePage="search">
      <form className="feature-form feature-card" onSubmit={submit}>
        <label>
          Search for
          <input
            value={term}
            onChange={(e) => setTerm(e.target.value)}
            maxLength={100}
            required
          />
        </label>
        <label>
          Category
          <select value={kind} onChange={(e) => setKind(e.target.value)}>
            <option value="users">People</option>
            <option value="posts">Posts</option>
            <option value="groups">Groups</option>
          </select>
        </label>
        <button className="primary-button" disabled={busy}>
          Search
        </button>
      </form>
      {error && <p role="alert">{error}</p>}
      <div aria-live="polite">
        {items.map((item) => (
          <article key={item.id} className="feature-result">
            {shown.kind === "users" ? (
              <Link href={"/profile/" + item.id}>{item.title}</Link>
            ) : shown.kind === "groups" ? (
              <>
                <h2>{item.title}</h2>
                <p>{item.description}</p>
                <Link
                  href={
                    "/groups/discover?search=" +
                    encodeURIComponent(item.title || "")
                  }
                >
                  Find group in Discover
                </Link>
              </>
            ) : (
              <>
                <Link href={"/profile/" + item.userId}>View author</Link>
                <p>{item.content}</p>
                <LikeButton
                  id={item.id}
                  count={item.likeCount || 0}
                  liked={item.isLiked || false}
                />
                <ContentActions kind="post" id={item.id} />
              </>
            )}
          </article>
        ))}
        {shown.term && !busy && !items.length && <p>No results.</p>}
      </div>
      {more && (
        <button disabled={busy} onClick={() => run(query.current.page + 1)}>
          Load more results
        </button>
      )}
    </FeatureShell>
  );
}
