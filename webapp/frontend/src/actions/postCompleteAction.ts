import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { Action, AnyAction } from 'redux';
import { CompleteReq, CompleteRes } from '../types/appApiTypes';
import { fetchItemAction } from './fetchItemAction';

export const POST_COMPLETE_START = 'POST_COMPLETE_START';
export const POST_COMPLETE_SUCCESS = 'POST_COMPLETE_SUCCESS';
export const POST_COMPLETE_FAIL = 'POST_COMPLETE_FAIL';

type ThunkResult<R> = ThunkAction<R, void, undefined, AnyAction>;

export function postCompleteAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
    Promise.resolve()
      .then(() => {
        dispatch(postCompleteStartAction());
      })
      .then(() => {
        return AppClient.post('/complete', {
          item_id: itemId,
        } as CompleteReq);
      })
      .then((response: Response) => {
        if (response.status !== 200) {
          throw new Error('HTTP status not 200');
        }

        return response.json();
      })
      .then((body: CompleteRes) => {
        dispatch(postCompleteSuccessAction());
        dispatch(fetchItemAction(itemId.toString())); // FIXME: 異常系のハンドリングが取引ページ向けでない
      })
      .catch((err: Error) => {
        dispatch(
          postCompleteFailAction({
            error: err.message,
          }),
        );
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
  extends Action<typeof POST_COMPLETE_FAIL> {
  payload: FormErrorState;
}

export function postCompleteFailAction(
  newErros: FormErrorState,
): PostCompleteFailAction {
  return {
    type: POST_COMPLETE_FAIL,
    payload: newErros,
  };
}
