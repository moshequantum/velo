function timingSafeEqual(a: string, b: string): boolean {
  const enc = new TextEncoder();
  const aBuf = enc.encode(a);
  const bBuf = enc.encode(b);
  if (aBuf.byteLength !== bBuf.byteLength) {
    return false;
  }
  return crypto.subtle.timingSafeEqual(aBuf, bBuf);
}

export async function POST(request: Request) {
  const secret = request.headers.get('X-Telegram-Bot-Api-Secret-Token');
  const expectedSecret = process.env.TELEGRAM_WEBHOOK_SECRET;

  if (!expectedSecret) {
    return Response.json({ error: 'Webhook secret unconfigured' }, { status: 503 });
  }

  if (!secret || !timingSafeEqual(secret, expectedSecret)) {
    return Response.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const payload = (await request.json().catch(() => null)) as Record<string, any> | null;
  if (!payload) {
    return Response.json({ error: 'Invalid JSON' }, { status: 400 });
  }

  const expectedOwner = process.env.TELEGRAM_OWNER_ID;
  const fromId = String(payload.message?.from?.id ?? payload.callback_query?.from?.id ?? '');
  if (expectedOwner && fromId !== String(expectedOwner)) {
    return Response.json({ error: 'Forbidden' }, { status: 403 });
  }

  if (payload.callback_query?.data) {
    const action = String(payload.callback_query.data);
    const match = action.match(/^(approve|reject):(.+)$/);
    if (!match) {
      return Response.json({ error: 'Invalid callback action format' }, { status: 400 });
    }
    const [, decision, actionId] = match;
    console.log('[VELO HITL] Action ' + actionId + ': decision ' + decision);
  }

  return Response.json({ ok: true });
}
