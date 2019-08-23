import { ItemData } from '../dataObjects/item';
import { FETCH_ITEM_SUCCESS } from '../actions/fetchItemAction';
import { ActionTypes } from '../actions/actionTypes';

export interface ViewingItemState {
  item?: ItemData;
}

const initialState: ViewingItemState = {};

const viewingItem = (
  state: ViewingItemState = initialState,
  action: ActionTypes,
): ViewingItemState => {
  switch (action.type) {
    case FETCH_ITEM_SUCCESS:
      return { ...state, item: action.payload.item };
    default:
      return state;
  }
};

export default viewingItem;
