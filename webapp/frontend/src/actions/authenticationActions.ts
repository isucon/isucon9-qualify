import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from "redux-thunk";
import { FormErrorState } from "../reducers/formErrorReducer";
import { push } from 'connected-react-router';
import {AnyAction} from "redux";
import {routes} from "../routes/Route";
import {AppState} from "../index";
import {LoginRes} from "../types/appApiTypes";

export const LOGIN_SUCCESS = 'LOGIN_SUCCESS';
export const LOGIN_FAIL = 'LOGIN_FAIL';

type ThunkResult<R> = ThunkAction<R, AppState, undefined, AnyAction>

export function postLoginAction(accountName: string, password: string): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
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
            .then((body: LoginRes) => {
                dispatch(loginSuccessAction({
                    userId: body.id,
                    accountName: body.account_name,
                    address: body.address,
                }));
                dispatch(push(routes.top.path))
            })
            .catch((err: Error) => {
                dispatch(loginFailAction({
                    error: err.message,
                }))
            })
    };
}

export interface LoginSuccessAction {
    type: typeof LOGIN_SUCCESS,
    payload: {
        userId: number,
        accountName: string,
        address?: string,
    },
}

export function loginSuccessAction(newAuthState: { userId: number, accountName: string, address?: string }): LoginSuccessAction {
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
