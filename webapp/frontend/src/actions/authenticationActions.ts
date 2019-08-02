import { AuthStatusState } from "../reducers/authStatusReducer";
import AppClient from '../httpClients/appClient';
import {ThunkAction, ThunkDispatch} from "redux-thunk";
import {FormErrorState} from "../reducers/formErrorReducer";

export const LOGIN_SUCCESS = 'LOGIN_SUCCESS';
export const LOGIN_FAIL = 'LOGIN_FAIL';

type State = void | AuthStatusState;
type ActionTypes = LoginSuccessAction | LoginFailAction;
type ThunkResult<R> = ThunkAction<R, State, undefined, ActionTypes>

export function postLoginAction(accountName: string, password: string): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, ActionTypes>, getState: () => any) => {
        AppClient.post('/login', {
            account_name: accountName,
            password: password,
        })
            .then((response: Response) => {
                if (response.status !== 200) {
                    throw new Error('HTTP status not 200');
                }

                return response.json();
            })
            .then((body) => {
                dispatch(loginSuccessAction({
                    userId: body.id,
                    accountName: body.account_name,
                    address: body.address,
                }));
            })
            .catch((err: Error) => {
                dispatch(loginFailAction({
                    errorMsg: [err.message]
                }))
            })
    };
}

export interface LoginSuccessAction {
    type: typeof LOGIN_SUCCESS,
    payload: AuthStatusState,
}

export function loginSuccessAction(newAuthState: AuthStatusState): LoginSuccessAction {
    return {
        type: LOGIN_SUCCESS,
        payload: newAuthState,
    };
}

export interface LoginFailAction {
    type: typeof LOGIN_FAIL,
    payload: FormErrorState,
}

export function loginFailAction(newErros: FormErrorState): LoginFailAction {
    return {
        type: LOGIN_FAIL,
        payload: newErros,
    };
}
