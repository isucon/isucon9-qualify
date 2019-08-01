import { AuthStatusState } from "../reducers/authStatusReducer";
import AppClient from '../httpClients/appClient';
import {ThunkAction, ThunkDispatch} from "redux-thunk";

export const LOGIN_SUCCESS = 'LOGIN_SUCCESS';
export const LOGIN_FAIL = 'LOGIN_FAIL';

type State = void | AuthStatusState;
type ActionTypes = LoginSuccessAction | LoginFailAction;
type ThunkResult<R> = ThunkAction<R, State, undefined, ActionTypes>

export function postLoginAction(accountName: string, password: string): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, ActionTypes>, getState: () => any) => {
        AppClient.post('/login')
            .then((response: Response) => {
                if (response.status === 200) {
                    dispatch(loginSuccessAction({
                        userId: 1235, // TODO
                        accountName: 'sota1235', // TODO
                    }));
                }

                dispatch(loginFailAction())
            })
    }
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
}

export function loginFailAction(): LoginFailAction {
    return {
        type: LOGIN_FAIL,
    };
}
