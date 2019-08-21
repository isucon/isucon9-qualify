import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { FormErrorState } from '../reducers/formErrorReducer';
import { Action, AnyAction } from 'redux';
import { ErrorRes, BumpReq, BumpRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';

export const POST_BUMP_START = 'POST_BUMP_START';
export const POST_BUMP_SUCCESS = 'POST_BUMP_SUCCESS';
export const POST_BUMP_FAIL = 'POST_BUMP_FAIL';

type ThunkResult<R> = ThunkAction<R, void, undefined, AnyAction>;

export function postBumpAction(itemId: number): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
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
        dispatch(
          postBumpFailAction({
            error: err.message, // TODO
          }),
        );
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
  extends Action<typeof POST_BUMP_SUCCESS> {}

export function postBumpSuccessAction(): PostBumpSuccessAction {
  return {
    type: POST_BUMP_SUCCESS,
  };
}

export interface PostBumpFailAction extends Action<typeof POST_BUMP_FAIL> {
  payload: FormErrorState;
}

export function postBumpFailAction(
  newErrors: FormErrorState,
): PostBumpFailAction {
  return {
    type: POST_BUMP_FAIL,
    payload: newErrors,
  };
}
