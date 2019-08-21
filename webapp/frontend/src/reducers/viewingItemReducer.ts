import { AnyAction } from 'redux';
import { ItemData } from '../dataObjects/item';
import {
  FETCH_ITEM_SUCCESS,
  FetchItemSuccessAction,
} from '../actions/fetchItemAction';

export interface ViewingItemState {
  item?: ItemData;
}

const initialState: ViewingItemState = {};

type actions = AnyAction | FetchItemSuccessAction;

const viewingItem = (
  state: ViewingItemState = initialState,
  action: actions,
): ViewingItemState => {
  switch (action.type) {
    case FETCH_ITEM_SUCCESS:
      return { ...state, item: action.payload.item };
    default:
      return state;
  }
};

export default viewingItem;
