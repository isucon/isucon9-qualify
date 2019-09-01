import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { ErrorRes, ShipDoneReq, ShipDoneRes } from '../types/appApiTypes';
import { fetchItemAction } from './fetchItemAction';
import { AppResponseError } from '../errors/AppResponseError';
import { AppState } from '../index';
import { SnackBarAction } from './actionTypes';

export const POST_SHIPPED_DONE_START = 'POST_SHIPPED_DONE_START';
export const POST_SHIPPED_DONE_SUCCESS = 'POST_SHIPPED_DONE_SUCCESS';
export const POST_SHIPPED_DONE_FAIL = 'POST_SHIPPED_DONE_FAIL';
export type PostShippedDoneActions =
  | PostShippedDoneStartAction
  | PostShippedDoneSuccessAction
  | PostShippedDoneFailAction;
type ThunkResult<R> = ThunkAction<
  R,
  AppState,
  undefined,
  PostShippedDoneActions
>;

export function postShippedDoneAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, PostShippedDoneActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(postShippedDoneStartAction());
      })
      .then(() => {
        return AppClient.post('/ship_done', {
          item_id: itemId,
        } as ShipDoneReq);
      })
      .then(async (response: Response) => {
        if (response.status !== 200) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

        return await response.json();
      })
      .then((body: ShipDoneRes) => {
        dispatch(postShippedDoneSuccessAction());
      })
      .then(() => {
        dispatch(fetchItemAction(itemId.toString())); // FIXME: 異常系のハンドリングが取引ページ向けでない
      })
      .catch((err: Error) => {
        dispatch(postShippedDoneFailAction(err.message));
      });
  };
}

export interface PostShippedDoneStartAction
  extends Action<typeof POST_SHIPPED_DONE_START> {}

export function postShippedDoneStartAction(): PostShippedDoneStartAction {
  return {
    type: POST_SHIPPED_DONE_START,
  };
}

export interface PostShippedDoneSuccessAction
  extends Action<typeof POST_SHIPPED_DONE_SUCCESS> {}

export function postShippedDoneSuccessAction(): PostShippedDoneSuccessAction {
  return {
    type: POST_SHIPPED_DONE_SUCCESS,
  };
}

export interface PostShippedDoneFailAction
  extends SnackBarAction<typeof POST_SHIPPED_DONE_FAIL> {}

export function postShippedDoneFailAction(
  error: string,
): PostShippedDoneFailAction {
  return {
    type: POST_SHIPPED_DONE_FAIL,
    snackBarMessage: error,
    variant: 'error',
  };
}
