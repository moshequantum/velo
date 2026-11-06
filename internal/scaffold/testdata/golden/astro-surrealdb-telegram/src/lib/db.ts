import { Surreal } from 'surrealdb';

export interface PendingAction {
  id?: string;
  kind: string;
  payload: Record<string, unknown>;
  status: 'pending' | 'approved' | 'rejected' | 'expired';
  created_at?: string;
  expires_at?: string;
}

export interface DbConfig {
  url?: string;
  namespace?: string;
  database?: string;
  username?: string;
  password?: string;
}

/**
 * Creates and initializes a SurrealDB client for edge environments using HTTP engine.
 */
export async function createDbClient(config?: DbConfig): Promise<Surreal> {
  const db = new Surreal();

  const url = config?.url ?? (typeof process !== 'undefined' ? process.env?.SURREAL_URL : undefined) ?? 'http://127.0.0.1:8000';
  const namespace = config?.namespace ?? (typeof process !== 'undefined' ? process.env?.SURREAL_NS : undefined) ?? 'test';
  const database = config?.database ?? (typeof process !== 'undefined' ? process.env?.SURREAL_DB : undefined) ?? 'test';
  const username = config?.username ?? (typeof process !== 'undefined' ? process.env?.SURREAL_USER : undefined) ?? 'root';
  const password = config?.password ?? (typeof process !== 'undefined' ? process.env?.SURREAL_PASS : undefined) ?? 'root';

  await db.connect(url);
  await db.use({ namespace, database });

  if (username && password) {
    await db.signin({ username, password });
  }

  return db;
}
