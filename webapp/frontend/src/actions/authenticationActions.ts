import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { CallHistoryMethodAction, push } from 'connected-react-router';
import { routes } from '../routes/Route';
import { AppState } from '../index';
import { ErrorRes, LoginRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';
import { SnackBarAction } from './actionTypes';

export const LOGIN_SUCCESS = 'LOGIN_SUCCESS';
export const LOGIN_FAIL = 'LOGIN_FAIL';

export type AuthActions =
  | LoginSuccessAction
  | LoginFailAction
  | CallHistoryMethodAction;

type ThunkResult<R> = ThunkAction<R, AppState, undefined, AuthActions>;

export function postLoginAction(
  accountName: string,
  password: string,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, AuthActions>) => {
    AppClient.post(
      '/login',
      {
        account_name: accountName,
        password: password,
      },
      false,
    )
      .then(async (response: Response) => {
        if (response.status !== 200) {
          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

        return await response.json();
      })
      .then((body: LoginRes) => {
        dispatch(
          loginSuccessAction({
            userId: body.id,
            accountName: body.account_name,
            address: body.address,
          }),
        );
        dispatch(push(routes.top.path));
      })
      .catch((err: Error) => {
        dispatch(loginFailAction(err.message));
      });
  };
}

export interface LoginSuccessAction {
  type: typeof LOGIN_SUCCESS;
  payload: {
    userId: number;
    accountName: string;
    address?: string;
  };
}

export function loginSuccessAction(newAuthState: {
  userId: number;
  accountName: string;
  address?: string;
}): LoginSuccessAction {
  return {
    type: LOGIN_SUCCESS,
    payload: newAuthState,
  };
}

export interface LoginFailAction extends SnackBarAction<typeof LOGIN_FAIL> {}

export function loginFailAction(error: string): LoginFailAction {
  return {
    type: LOGIN_FAIL,
    snackBarMessage: error,
    variant: 'error',
  };
}
