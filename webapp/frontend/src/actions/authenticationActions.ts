import {AuthStatusState} from "../reducers/authStatusReducer";

export const POST_LOGIN = 'POST_LOGIN';
export const LOGIN_SUCCESS = 'LOGIN_SUCCESS';

export interface PostLoginAction {
    type: typeof POST_LOGIN,
    accountName: string,
    password: string,
}

export function postLoginAction(accountName: string, password: string): PostLoginAction {
    return { type: POST_LOGIN, accountName, password };
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
