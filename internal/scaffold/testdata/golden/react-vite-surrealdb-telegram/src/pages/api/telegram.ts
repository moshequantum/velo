import type { APIRoute } from 'astro';

function timingSafeEqual(a: string, b: string): boolean {
  const enc = new TextEncoder();
  const aBuf = enc.encode(a);
  const bBuf = enc.encode(b);
  if (aBuf.byteLength !== bBuf.byteLength) {
    return false;
  }
  return crypto.subtle.timingSafeEqual(aBuf, bBuf);
}

export const POST: APIRoute = async ({ request }) => {
  const secret = request.headers.get('X-Telegram-Bot-Api-Secret-Token');
  const expectedSecret = import.meta.env.TELEGRAM_WEBHOOK_SECRET;

  // Fail-closed: si no hay secreto configurado en el entorno -> 503
  if (!expectedSecret) {
    return new Response(JSON.stringify({ error: 'Webhook secret unconfigured' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // Comparación constant-time; si falta el token o no coincide -> 401
  if (!secret || !timingSafeEqual(secret, expectedSecret)) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const payload = await request.json().catch(() => null);
  if (!payload) {
    return new Response(JSON.stringify({ error: 'Invalid JSON' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // Validación de owner TELEGRAM_OWNER_ID -> 403
  const expectedOwner = import.meta.env.TELEGRAM_OWNER_ID;
  const fromId = String(payload.message?.from?.id ?? payload.callback_query?.from?.id ?? '');
  if (expectedOwner && fromId !== String(expectedOwner)) {
    return new Response(JSON.stringify({ error: 'Forbidden' }), {
      status: 403,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // HITL callback_query con formato approve:<id> / reject:<id>
  if (payload.callback_query?.data) {
    const action = String(payload.callback_query.data);
    const match = action.match(/^(approve|reject):(.+)$/);
    if (!match) {
      return new Response(JSON.stringify({ error: 'Invalid callback action format' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }
    const [, decision, actionId] = match;
    console.log('[VELO HITL] Action ' + actionId + ': decision ' + decision);
  }

  return new Response(JSON.stringify({ ok: true }), {
    headers: { 'Content-Type': 'application/json' },
  });
};
