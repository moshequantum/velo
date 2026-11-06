import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Velo Edge Application',
  description: 'Deployed on Cloudflare Workers with OpenNext',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="es">
      <body className="bg-slate-950 text-white min-h-screen">
        {children}
      </body>
    </html>
  );
}
