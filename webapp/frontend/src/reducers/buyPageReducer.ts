import {
  BUY_FAIL,
  BUY_START,
  BUY_SUCCESS,
  USING_CARD_FAIL,
} from '../actions/buyAction';
import { ActionTypes } from '../actions/actionTypes';

export interface BuyPageState {
  loadingBuy: boolean;
}

const initialState: BuyPageState = {
  loadingBuy: false,
};

const buyPage = (
  state: BuyPageState = initialState,
  action: ActionTypes,
): BuyPageState => {
  switch (action.type) {
    case BUY_START:
      return { ...state, loadingBuy: true };
    case BUY_SUCCESS:
    case BUY_FAIL:
    case USING_CARD_FAIL:
      return { ...state, loadingBuy: false };
    default:
      return { ...state };
  }
};

export default buyPage;
