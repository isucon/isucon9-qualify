import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { Action, AnyAction } from 'redux';
import { ErrorRes, ShipDoneReq, ShipDoneRes } from '../types/appApiTypes';
import { fetchItemAction } from './fetchItemAction';
import { AppResponseError } from '../errors/AppResponseError';

export const POST_SHIPPED_DONE_START = 'POST_SHIPPED_DONE_START';
export const POST_SHIPPED_DONE_SUCCESS = 'POST_SHIPPED_DONE_SUCCESS';
export const POST_SHIPPED_DONE_FAIL = 'POST_SHIPPED_DONE_FAIL';

type ThunkResult<R> = ThunkAction<R, void, undefined, AnyAction>;

export function postShippedDoneAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
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
        dispatch(
          postShippedDoneFailAction({
            error: err.message,
          }),
        );
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
  extends Action<typeof POST_SHIPPED_DONE_FAIL> {
  payload: FormErrorState;
}

export function postShippedDoneFailAction(
  newErrors: FormErrorState,
): PostShippedDoneFailAction {
  return {
    type: POST_SHIPPED_DONE_FAIL,
    payload: newErrors,
  };
}
