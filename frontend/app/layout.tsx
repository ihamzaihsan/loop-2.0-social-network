import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import "./v2.css";
import RealtimeAlerts from "../components/RealtimeAlerts";
import ThemeToggle from "../components/ThemeToggle";
import "./theme.css";
import "./features.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Loop — Your people, your pace",
  description: "Share moments, build communities, and stay close to your people.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  interactiveWidget: "resizes-content",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" data-scroll-behavior="smooth" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: `(function(){var p;try{p=localStorage.getItem('loop-theme')}catch(e){}document.documentElement.dataset.theme=p==='dark'||p==='light'?p:matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'})()` }} />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
        <RealtimeAlerts />
        <ThemeToggle authOnly />
      </body>
    </html>
  );
}
