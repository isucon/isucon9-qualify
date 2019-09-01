import { NotFoundError } from '../errors/NotFoundError';
import {
  internalServerError,
  InternalServerErrorAction,
  notFoundError,
  NotFoundErrorAction,
} from '../actions/errorAction';
import { InternalServerError } from '../errors/InternalServerError';
import { Action } from 'redux';

export function ajaxErrorHandler<T extends Action<any>>(
  err: Error,
  actionCreate: (message: string) => T,
): T | NotFoundErrorAction | InternalServerErrorAction {
  if (err instanceof NotFoundError) {
    return notFoundError(err.message);
  }

  if (err instanceof InternalServerError) {
    return internalServerError(err.message);
  }

  return actionCreate(err.message);
}
