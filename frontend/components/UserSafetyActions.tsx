"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import ActionDialog from "./ActionDialog";
import ContentActions from "./ContentActions";
import { api, refreshResources } from "../utils/api";

export default function UserSafetyActions({
  id,
  name,
}: {
  id: number;
  name: string;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const block = async () => {
    setBusy(true);
    setError("");
    try {
      await api(`/blocks?id=${id}`, { method: "POST" });
      refreshResources("social", "posts", "profiles", "users", "chat");
      router.push("/settings?tab=blocked");
    } catch (error) {
      setError(
        error instanceof Error ? error.message : "Unable to block user.",
      );
      setBusy(false);
    }
  };
  return (
    <div className="feature-actions">
      <button type="button" onClick={() => setOpen(true)}>
        Block user
      </button>
      <ContentActions kind="user" id={id} />
      {open && (
        <ActionDialog
          title={`Block ${name}`}
          onClose={() => {
            if (!busy) setOpen(false);
          }}
        >
          <p>
            Blocking removes follows and prevents direct messaging and access to
            each other’s profile and feed posts. Shared group conversations
            remain visible to group members.
          </p>
          {error && <p role="alert">{error}</p>}
          <button className="primary-button" disabled={busy} onClick={block}>
            {busy ? "Blocking…" : "Confirm block"}
          </button>
        </ActionDialog>
      )}
    </div>
  );
}
