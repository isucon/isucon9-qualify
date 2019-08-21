import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { Action, AnyAction } from 'redux';
import { ErrorRes, ShipReq, ShipRes } from '../types/appApiTypes';
import { fetchItemAction } from './fetchItemAction';
import { AppResponseError } from '../errors/AppResponseError';

export const POST_SHIPPED_START = 'POST_SHIPPED_START';
export const POST_SHIPPED_SUCCESS = 'POST_SHIPPED_SUCCESS';
export const POST_SHIPPED_FAIL = 'POST_SHIPPED_FAIL';

type ThunkResult<R> = ThunkAction<R, void, undefined, AnyAction>;

export function postShippedAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
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
        dispatch(
          postShippedFailAction({
            error: err.message,
          }),
        );
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
  extends Action<typeof POST_SHIPPED_FAIL> {
  payload: FormErrorState;
}

export function postShippedFailAction(
  newErrors: FormErrorState,
): PostShippedFailAction {
  return {
    type: POST_SHIPPED_FAIL,
    payload: newErrors,
  };
}
