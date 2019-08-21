import { LOGIN_FAIL, LoginFailAction } from '../actions/authenticationActions';
import { AnyAction } from 'redux';
import { REGISTER_FAIL, RegisterFailAction } from '../actions/registerAction';
import {
  SELLING_ITEM_FAIL,
  SellingFailAction,
} from '../actions/sellingItemAction';
import {
  BUY_FAIL,
  BuyFailAction,
  USING_CARD_FAIL,
  UsingCardFailAction,
} from '../actions/buyAction';
import {
  POST_ITEM_EDIT_FAIL,
  PostItemEditFailAction,
} from '../actions/postItemEditAction';

export interface FormErrorState {
  error?: string;
  buyFormError?: BuyFormErrorState;
  itemEditFormError?: string;
}

export interface BuyFormErrorState {
  cardError?: string; // Error from payment service
  buyError?: string; // Error from app
}

const initialState: FormErrorState = {
  error: undefined,
  buyFormError: {},
  itemEditFormError: undefined,
};

type Actions =
  | AnyAction
  | PostItemEditFailAction
  | LoginFailAction
  | RegisterFailAction
  | UsingCardFailAction
  | BuyFailAction
  | SellingFailAction;

const formError = (
  state: FormErrorState = initialState,
  action: Actions,
): FormErrorState => {
  switch (action.type) {
    case LOGIN_FAIL:
    case REGISTER_FAIL:
    case USING_CARD_FAIL:
    case BUY_FAIL:
    case SELLING_ITEM_FAIL: {
      return {
        ...action.payload,
      };
    }
    case POST_ITEM_EDIT_FAIL:
      return {
        ...state,
        itemEditFormError: action.payload.itemEditFormError,
      };
    default:
      return initialState;
  }
};

export default formError;
