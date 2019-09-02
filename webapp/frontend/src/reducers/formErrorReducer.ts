import { SELLING_ITEM_FAIL } from '../actions/sellingItemAction';
import { BUY_FAIL, USING_CARD_FAIL } from '../actions/buyAction';
import { POST_ITEM_EDIT_FAIL } from '../actions/postItemEditAction';
import { ActionTypes } from '../actions/actionTypes';

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

const formError = (
  state: FormErrorState = initialState,
  action: ActionTypes,
): FormErrorState => {
  switch (action.type) {
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
