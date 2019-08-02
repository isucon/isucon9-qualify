import {LOGIN_SUCCESS} from "../actions/authenticationActions";
import {REGISTER_SUCCESS} from "../actions/registerAction";
import {AnyAction} from "redux";


export interface AuthStatusState {
    userId?: number
    accountName?: string
    address?: string,
}

const authStatus = (state: AuthStatusState = {}, action: AnyAction): AuthStatusState => {
    switch (action.type) {
        case LOGIN_SUCCESS:
        case REGISTER_SUCCESS: {
            return {
                ...state,
                ...action.payload,
            }
        }
        default:
            return state;
    }
};

export default authStatus;