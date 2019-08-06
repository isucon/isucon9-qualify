import {LOGIN_FAIL} from "../actions/authenticationActions";
import { AnyAction } from "redux";
import {REGISTER_FAIL} from "../actions/registerAction";
import {SELLING_ITEM_FAIL} from "../actions/sellingItemAction";
import {BUY_FAIL, USING_CARD_FAIL} from "../actions/buyAction";

export interface FormErrorState {
    error?: string,
    buyFormError?: BuyFormErrorState,
}

export interface BuyFormErrorState {
    cardError?: string, // Error from payment service
    buyError?: string,  // Error from app
}

const initialState: FormErrorState = {
    error: undefined,
    buyFormError: {},
};

const formError = (state: FormErrorState = initialState, action: AnyAction): FormErrorState => {
    switch (action.type) {
        case LOGIN_FAIL:
        case REGISTER_FAIL:
        case USING_CARD_FAIL:
        case BUY_FAIL:
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