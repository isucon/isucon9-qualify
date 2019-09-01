import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { CompleteReq, CompleteRes, ErrorRes } from '../types/appApiTypes';
import { fetchItemAction } from './fetchItemAction';
import { AppResponseError } from '../errors/AppResponseError';
import { AppState } from '../index';
import { SnackBarAction } from './actionTypes';

export const POST_COMPLETE_START = 'POST_COMPLETE_START';
export const POST_COMPLETE_SUCCESS = 'POST_COMPLETE_SUCCESS';
export const POST_COMPLETE_FAIL = 'POST_COMPLETE_FAIL';

export type PostCompleteActions =
  | PostCompleteStartAction
  | PostCompleteSuccessAction
  | PostCompleteFailAction;
type ThunkResult<R> = ThunkAction<R, AppState, undefined, PostCompleteActions>;

export function postCompleteAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, PostCompleteActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(postCompleteStartAction());
      })
      .then(() => {
        return AppClient.post('/complete', {
          item_id: itemId,
        } as CompleteReq);
      })
      .then(async (response: Response) => {
        if (response.status !== 200) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

        return await response.json();
      })
      .then((body: CompleteRes) => {
        dispatch(postCompleteSuccessAction());
      })
      .then(() => {
        dispatch(fetchItemAction(itemId.toString())); // FIXME: 異常系のハンドリングが取引ページ向けでない
      })
      .catch((err: Error) => {
        dispatch(postCompleteFailAction(err.message));
      });
  };
}

export interface PostCompleteStartAction
  extends Action<typeof POST_COMPLETE_START> {}

export function postCompleteStartAction(): PostCompleteStartAction {
  return {
    type: POST_COMPLETE_START,
  };
}

export interface PostCompleteSuccessAction
  extends Action<typeof POST_COMPLETE_SUCCESS> {}

export function postCompleteSuccessAction(): PostCompleteSuccessAction {
  return {
    type: POST_COMPLETE_SUCCESS,
  };
}

export interface PostCompleteFailAction
  extends SnackBarAction<typeof POST_COMPLETE_FAIL> {}

export function postCompleteFailAction(error: string): PostCompleteFailAction {
  return {
    type: POST_COMPLETE_FAIL,
    snackBarMessage: error,
    variant: 'error',
  };
}
