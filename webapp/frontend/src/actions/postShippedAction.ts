import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { ErrorRes, ShipReq, ShipRes } from '../types/appApiTypes';
import { fetchItemAction } from './fetchItemAction';
import { AppResponseError } from '../errors/AppResponseError';
import { AppState } from '../index';
import { SnackBarAction } from './actionTypes';

export const POST_SHIPPED_START = 'POST_SHIPPED_START';
export const POST_SHIPPED_SUCCESS = 'POST_SHIPPED_SUCCESS';
export const POST_SHIPPED_FAIL = 'POST_SHIPPED_FAIL';

export type PostShippedActions =
  | PostShippedStartAction
  | PostShippedSuccessAction
  | PostShippedFailAction;
type ThunkResult<R> = ThunkAction<R, AppState, undefined, PostShippedActions>;

export function postShippedAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, PostShippedActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(postShippedStartAction());
      })
      .then(() => {
        return AppClient.post('/ship', {
          item_id: itemId,
        } as ShipReq);
      })
      .then(async (response: Response) => {
        if (response.status !== 200) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

        return await response.json();
      })
      .then((body: ShipRes) => {
        dispatch(postShippedSuccessAction());
      })
      .then(() => {
        dispatch(fetchItemAction(itemId.toString())); // FIXME: 異常系のハンドリングが取引ページ向けでない
      })
      .catch((err: Error) => {
        dispatch(postShippedFailAction(err.message));
      });
  };
}

export interface PostShippedStartAction
  extends Action<typeof POST_SHIPPED_START> {}

export function postShippedStartAction(): PostShippedStartAction {
  return {
    type: POST_SHIPPED_START,
  };
}

export interface PostShippedSuccessAction
  extends Action<typeof POST_SHIPPED_SUCCESS> {}

export function postShippedSuccessAction(): PostShippedSuccessAction {
  return {
    type: POST_SHIPPED_SUCCESS,
  };
}

export interface PostShippedFailAction
  extends SnackBarAction<typeof POST_SHIPPED_FAIL> {}

export function postShippedFailAction(error: string): PostShippedFailAction {
  return {
    type: POST_SHIPPED_FAIL,
    snackBarMessage: error,
    variant: 'error',
  };
}
