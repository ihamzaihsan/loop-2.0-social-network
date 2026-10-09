"use client";

import { ReactNode, useEffect } from "react";
import { useRouter } from "next/navigation";
import Sidebar from "./Sidebar";
import { redirectBasedOnSession } from "../utils/session";

export default function FeatureShell({
  title,
  children,
  activePage,
}: {
  title: string;
  children: ReactNode;
  activePage: string;
}) {
  const router = useRouter();
  useEffect(() => {
    redirectBasedOnSession(router, true);
  }, [router]);
  return (
    <div>
      <Sidebar activePage={activePage} />
      <main className="main-content">
        <div className="feature-page">
          <h1>{title}</h1>
          {children}
        </div>
      </main>
    </div>
  );
}
