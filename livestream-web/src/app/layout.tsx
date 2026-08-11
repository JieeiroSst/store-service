import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { getSession } from "@/lib/session";
import Nav from "@/components/Nav";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });

export const metadata: Metadata = {
  title: "livestream",
  description: "Frontend for livestream_service",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const session = getSession();
  return (
    <html lang="en" className={inter.variable}>
      <body className="font-sans">
        <Nav session={session} />
        <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">{children}</main>
      </body>
    </html>
  );
}
