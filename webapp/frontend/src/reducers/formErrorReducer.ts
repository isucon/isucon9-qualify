import {LOGIN_FAIL} from "../actions/authenticationActions";
import { AnyAction } from "redux";
import {REGISTER_FAIL} from "../actions/registerAction";
import {SELLING_ITEM_FAIL} from "../actions/sellingItemAction";

export interface FormErrorState {
    errorMsg: string[],
    buyFormError?: BuyFormErrorState,
}

export interface BuyFormErrorState {
    cardError: string[], // Error from payment service
    buyError: string[], // Error from app
}

const initialState: FormErrorState = {
    errorMsg: [],
    buyFormError: {
        cardError: [],
        buyError: [],
    }
};

const formError = (state: FormErrorState = initialState, action: AnyAction): FormErrorState => {
    switch (action.type) {
        case LOGIN_FAIL:
        case REGISTER_FAIL:
        case SELLING_ITEM_FAIL: {
            return {
                ...action.payload,
            }
        }
        default:
            return initialState;
    }
};

export default formError;