import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { CallHistoryMethodAction, push } from 'connected-react-router';
import { ErrorRes, RegisterReq, RegisterRes } from '../types/appApiTypes';
import { routes } from '../routes/Route';
import { AppResponseError } from '../errors/AppResponseError';
import { AppState } from '../index';
import { SnackBarAction } from './actionTypes';

export const REGISTER_SUCCESS = 'REGISTER_SUCCESS';
export const REGISTER_FAIL = 'REGISTER_FAIL';

export type RegisterActions =
  | RegisterSuccessAction
  | RegisterFailAction
  | CallHistoryMethodAction;
type ThunkResult<R> = ThunkAction<R, AppState, undefined, RegisterActions>;

export function postRegisterAction(payload: RegisterReq): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, RegisterActions>) => {
    AppClient.post('/register', payload, false)
      .then(async (response: Response) => {
        if (response.status !== 200) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

        return await response.json();
      })
      .then((body: RegisterRes) => {
        dispatch(
          registerSuccessAction({
            userId: body.id,
            accountName: body.account_name,
            address: body.address,
            numSellItems: body.num_sell_items,
          }),
        );
        dispatch(push(routes.top.path));
      })
      .catch((err: Error) => {
        dispatch(registerFailAction(err.message));
      });
  };
}

export interface RegisterSuccessAction extends Action<typeof REGISTER_SUCCESS> {
  payload: {
    userId: number;
    accountName: string;
    address: string;
  };
}

export function registerSuccessAction(newAuthState: {
  userId: number;
  accountName: string;
  address: string;
  numSellItems: number;
}): RegisterSuccessAction {
  return {
    type: REGISTER_SUCCESS,
    payload: newAuthState,
  };
}

export interface RegisterFailAction
  extends SnackBarAction<typeof REGISTER_FAIL> {}

export function registerFailAction(error: string): RegisterFailAction {
  return {
    type: REGISTER_FAIL,
    snackBarMessage: error,
    variant: 'error',
  };
}
