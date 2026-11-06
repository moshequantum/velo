import React from 'react';

export default function App() {
  return (
    <div className="min-h-screen flex items-center justify-center p-6 bg-slate-950 text-white">
      <main className="max-w-xl text-center space-y-4">
        <div className="inline-block bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 text-xs px-3 py-1 rounded-full uppercase tracking-wider font-semibold">
          Velo Framework Active
        </div>
        <h1 className="text-4xl font-extrabold tracking-tight">mi-app-vite</h1>
        <p className="text-slate-400">
          React-Vite SPA en Cloudflare Pages con funciones Edge y Cockpit Telegram.
        </p>
      </main>
    </div>
  );
}
