import { NotFoundError } from '../errors/NotFoundError';
import { internalServerError, notFoundError } from '../actions/errorAction';
import { InternalServerError } from '../errors/InternalServerError';
import { ActionTypes } from '../actions/actionTypes';

export function ajaxErrorHandler(
  err: Error,
  actionCreater: (message: string) => ActionTypes,
): ActionTypes {
  if (err instanceof NotFoundError) {
    return notFoundError(err.message);
  }

  if (err instanceof InternalServerError) {
    return internalServerError(err.message);
  }

  return actionCreater(err.message);
}
