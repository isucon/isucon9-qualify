import { randomInt } from 'crypto';
import bcrypt from 'bcrypt';
import { BCRYPT_COST } from './constants.js';

export function secureRandomStr(length: number): string {
  const k = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-';
  let result = '';

  for (let i = 0; i < length; i++) {
    result += k[randomInt(k.length)];
  }

  return result;
}

export async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, BCRYPT_COST);
}

export async function verifyPassword(password: string, hash: string | Buffer): Promise<boolean> {
  const hashStr = typeof hash === 'string' ? hash : hash.toString();
  return bcrypt.compare(password, hashStr);
}

export function getImageURL(imageName: string): string {
  return `/upload/${imageName}`;
}
