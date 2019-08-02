import { AuthStatusState } from "../reducers/authStatusReducer";
import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from "redux-thunk";
import { FormErrorState } from "../reducers/formErrorReducer";
import { push } from 'connected-react-router';
import {AnyAction} from "redux";
import {RegisterReq, RegisterRes} from "../types/appApiTypes";

export const REGISTER_SUCCESS = 'REGISTER_SUCCESS';
export const REGISTER_FAIL = 'REGISTER_FAIL';

type State = void | AuthStatusState;
type ThunkResult<R> = ThunkAction<R, State, undefined, AnyAction>

export function postRegisterAction(payload: RegisterReq): ThunkResult<void> {
    return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
        AppClient.post('/register', payload)
            .then((response: Response) => {
                if (response.status !== 200) {
                    throw new Error('HTTP status not 200');
                }

                return response.json();
            })
            .then((body: RegisterRes) => {
                dispatch(registerSuccessAction({
                    userId: body.id,
                    accountName: body.account_name,
                    address: body.address,
                }));
                dispatch(push('/items'))
            })
            .catch((err: Error) => {
                dispatch(registerFailAction({
                    errorMsg: [err.message]
                }))
            })
    };
}

export interface RegisterSuccessAction {
    type: typeof REGISTER_SUCCESS,
    payload: AuthStatusState,
}

export function registerSuccessAction(newAuthState: AuthStatusState): RegisterSuccessAction {
    return {
        type: REGISTER_SUCCESS,
        payload: newAuthState,
    };
}

export interface RegisterFailAction {
    type: typeof REGISTER_FAIL,
    payload: FormErrorState,
}

export function registerFailAction(newErros: FormErrorState): RegisterFailAction {
    return {
        type: REGISTER_FAIL,
        payload: newErros,
    };
}
