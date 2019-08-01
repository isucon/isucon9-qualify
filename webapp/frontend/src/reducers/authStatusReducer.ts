import { LOGIN_SUCCESS, LoginSuccessAction } from "../actions/authenticationActions";


export interface AuthStatusState {
    userId?: number
    accountName?: string
}

const authStatus = (state: AuthStatusState = {}, action: LoginSuccessAction): AuthStatusState => {
    switch (action.type) {
        case LOGIN_SUCCESS: {
            return {
                ...state,
                ...action.payload,
            }
        }
        default:
            return {};
    }
};

export default authStatus;