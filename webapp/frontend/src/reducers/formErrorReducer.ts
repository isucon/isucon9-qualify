import {LOGIN_FAIL, LOGIN_SUCCESS} from "../actions/authenticationActions";
import { AnyAction } from "redux";


export interface FormErrorState {
    errorMsg: string[],
}

const initialState: FormErrorState = {
    errorMsg: []
};

const formError = (state: FormErrorState = initialState, action: AnyAction): FormErrorState => {
    switch (action.type) {
        case LOGIN_FAIL: {
            return {
                ...action.payload,
            }
        }
        default:
            return initialState;
    }
};

export default formError;