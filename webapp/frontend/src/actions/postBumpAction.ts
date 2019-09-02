import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { ErrorRes, BumpReq, BumpRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';
import { AppState } from '../index';
import { SnackBarAction } from './actionTypes';

export const POST_BUMP_START = 'POST_BUMP_START';
export const POST_BUMP_SUCCESS = 'POST_BUMP_SUCCESS';
export const POST_BUMP_FAIL = 'POST_BUMP_FAIL';

export type PostBumpActions =
  | PostBumpStartAction
  | PostBumpSuccessAction
  | PostBumpFailAction;

type ThunkResult<R> = ThunkAction<R, AppState, undefined, PostBumpActions>;

export function postBumpAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, PostBumpActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(postBumpStartAction());
      })
      .then(() => {
        return AppClient.post('/bump', { item_id: itemId } as BumpReq);
      })
      .then(async (response: Response) => {
        if (response.status !== 200) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

        return await response.json();
      })
      .then((body: BumpRes) => {
        dispatch(postBumpSuccessAction());
      })
      .catch((err: Error) => {
        dispatch(postBumpFailAction(err.message));
      });
  };
}

export interface PostBumpStartAction extends Action<typeof POST_BUMP_START> {}

export function postBumpStartAction(): PostBumpStartAction {
  return {
    type: POST_BUMP_START,
  };
}

export interface PostBumpSuccessAction
  extends SnackBarAction<typeof POST_BUMP_SUCCESS> {}

export function postBumpSuccessAction(): PostBumpSuccessAction {
  return {
    type: POST_BUMP_SUCCESS,
    snackBarMessage: 'BUMPに成功しました',
    variant: 'success',
  };
}

export interface PostBumpFailAction
  extends SnackBarAction<typeof POST_BUMP_FAIL> {}

export function postBumpFailAction(error: string): PostBumpFailAction {
  return {
    type: POST_BUMP_FAIL,
    snackBarMessage: error,
    variant: 'error',
  };
}
