import { SessionOptions, getIronSession } from 'iron-session';
import { Context } from 'hono';
import { SESSION_NAME } from './constants.js';
import { secureRandomStr } from './utils.js';

export interface SessionData {
  userId?: number;
  csrfToken?: string;
}

const sessionOptions: SessionOptions = {
  password: process.env['SESSION_SECRET'] || secureRandomStr(32),
  cookieName: SESSION_NAME,
  cookieOptions: {
    secure: process.env['NODE_ENV'] === 'production',
    httpOnly: true,
    sameSite: 'lax',
  },
};

export async function getSession(c: Context): Promise<SessionData> {
  return getIronSession<SessionData>(c.req.raw, c.res, sessionOptions);
}

export async function saveSession(c: Context, session: SessionData): Promise<void> {
  const ironSession = await getIronSession<SessionData>(c.req.raw, c.res, sessionOptions);
  Object.assign(ironSession, session);
  await ironSession.save();
}

export async function destroySession(c: Context): Promise<void> {
  const session = await getIronSession<SessionData>(c.req.raw, c.res, sessionOptions);
  session.destroy();
}
