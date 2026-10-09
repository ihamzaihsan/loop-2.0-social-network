"use client";
import { FormEvent, useState } from "react";
import ActionDialog from "./ActionDialog";
import { useRouter } from "next/navigation";
import { api, refreshResources } from "../utils/api";
type Member = { user_id: number; first_name: string; last_name: string };
export default function GroupManagement({
  id,
  title,
  description,
  owner,
  members,
  currentId,
}: {
  id: number;
  title: string;
  description: string;
  owner: boolean;
  members: Member[];
  currentId: number;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function request(method: string, body?: object, exit = false) {
    setBusy(true);
    setError("");
    try {
      await api("/groups/manage?id=" + id, {
        method,
        body: body ? JSON.stringify(body) : undefined,
      });
      refreshResources("groups");
      if (exit) router.push("/groups");
      else setOpen(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to update group");
    } finally {
      setBusy(false);
    }
  }
  function edit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    request("PUT", Object.fromEntries(new FormData(e.currentTarget)));
  }
  return (
    <div className="group-management">
      <button className="primary-button" onClick={() => setOpen(true)}>
        {owner ? "Manage group" : "Leave group"}
      </button>
      {open && (
        <ActionDialog
          title={owner ? "Manage group" : "Leave group"}
          onClose={() => {
            if (!busy) setOpen(false);
          }}
        >
          {error && <p role="alert">{error}</p>}
          {owner ? (
            <>
              <form className="feature-form" onSubmit={edit}>
                <label>
                  Title
                  <input
                    name="title"
                    defaultValue={title}
                    maxLength={100}
                    required
                  />
                </label>
                <label>
                  Description
                  <textarea
                    name="description"
                    defaultValue={description}
                    maxLength={100}
                  />
                </label>
                <button disabled={busy}>Save group</button>
              </form>
              <h3>Members</h3>
              {members
                .filter((m) => m.user_id !== currentId)
                .map((m) => (
                  <div key={m.user_id} className="feature-result">
                    <span>
                      {m.first_name} {m.last_name}
                    </span>
                    <button
                      disabled={busy}
                      onClick={() => {
                        if (window.confirm("Remove this member?"))
                          request("POST", {
                            action: "remove",
                            userId: m.user_id,
                          });
                      }}
                    >
                      Remove member
                    </button>
                  </div>
                ))}
              <p>Deleting removes all group posts, events and chat history.</p>
              <button
                disabled={busy}
                onClick={() => {
                  if (window.confirm("Permanently delete this group?"))
                    request("DELETE", undefined, true);
                }}
              >
                Delete group
              </button>
            </>
          ) : (
            <>
              <p>You will lose access to this group until you join again.</p>
              <button
                disabled={busy}
                onClick={() => request("POST", { action: "leave" }, true)}
              >
                Confirm leave
              </button>
            </>
          )}
        </ActionDialog>
      )}
    </div>
  );
}
