function timingSafeEqual(a: string, b: string): boolean {
  const enc = new TextEncoder();
  const aBuf = enc.encode(a);
  const bBuf = enc.encode(b);
  if (aBuf.byteLength !== bBuf.byteLength) {
    return false;
  }
  return crypto.subtle.timingSafeEqual(aBuf, bBuf);
}

export const onRequestPost = async (context: {
  request: Request;
  env: { TELEGRAM_WEBHOOK_SECRET?: string; TELEGRAM_OWNER_ID?: string };
}) => {
  const secret = context.request.headers.get('X-Telegram-Bot-Api-Secret-Token');
  const expectedSecret = context.env.TELEGRAM_WEBHOOK_SECRET;

  if (!expectedSecret) {
    return new Response(JSON.stringify({ error: 'Webhook secret unconfigured' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  if (!secret || !timingSafeEqual(secret, expectedSecret)) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const payload = (await context.request.json().catch(() => null)) as Record<string, any> | null;
  if (!payload) {
    return new Response(JSON.stringify({ error: 'Invalid JSON' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const expectedOwner = context.env.TELEGRAM_OWNER_ID;
  const fromId = String(payload.message?.from?.id ?? payload.callback_query?.from?.id ?? '');
  if (expectedOwner && fromId !== String(expectedOwner)) {
    return new Response(JSON.stringify({ error: 'Forbidden' }), {
      status: 403,
      headers: { 'Content-Type': 'application/json' },
    });
  }

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
