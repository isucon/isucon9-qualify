import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { Action, AnyAction } from 'redux';
import { ShipReq, ShipRes } from '../types/appApiTypes';

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
      .then((response: Response) => {
        if (response.status !== 200) {
          throw new Error('HTTP status not 200');
        }

        return response.json();
      })
      .then((body: ShipRes) => {
        dispatch(postShippedSuccessAction());
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
