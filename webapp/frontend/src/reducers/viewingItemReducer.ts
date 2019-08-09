import { AnyAction } from 'redux';
import { ItemData } from '../dataObjects/item';
import {
  FETCH_ITEM_PAGE_SUCCESS,
  FetchItemPageSuccessAction,
} from '../actions/fetchItemPageAction';

export interface ViewingItemState {
  item?: ItemData;
}

const initialState: ViewingItemState = {};

type actions = AnyAction | FetchItemPageSuccessAction;

const viewingItem = (
  state: ViewingItemState = initialState,
  action: actions,
): ViewingItemState => {
  switch (action.type) {
    case FETCH_ITEM_PAGE_SUCCESS:
      return { ...state, item: action.payload.item };
    default:
      return state;
  }
};

export default viewingItem;
