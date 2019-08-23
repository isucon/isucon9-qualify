import { Action } from 'redux';

export const NOT_FOUND_ERROR = 'NOT_FOUND_ERROR';
export const INTERNAL_SERVER_ERROR = 'INTERNAL_SERVER_ERROR';

export type ErrorActions = NotFoundErrorAction | InternalServerErrorAction;

export interface NotFoundErrorAction extends Action<typeof NOT_FOUND_ERROR> {}

export function notFoundError(): NotFoundErrorAction {
  return { type: NOT_FOUND_ERROR };
}

export interface InternalServerErrorAction
  extends Action<typeof INTERNAL_SERVER_ERROR> {}

export function internalServerError(): InternalServerErrorAction {
  return { type: INTERNAL_SERVER_ERROR };
}
