import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Charon Admin',
  description: 'Operational console for Charon.',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
