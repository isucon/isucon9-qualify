import { TimelineItem } from '../dataObjects/item';
import { FETCH_USER_ITEMS_SUCCESS } from '../actions/fetchUserItemsAction';
import { FETCH_USER_PAGE_DATA_SUCCESS } from '../actions/fetchUserPageDataAction';
import { ActionTypes } from '../actions/actionTypes';
import { PATH_NAME_CHANGE } from '../actions/locationChangeAction';

export interface UserItemsState {
  items: TimelineItem[];
  hasNext: boolean;
}

const initialState: UserItemsState = {
  items: [],
  hasNext: false,
};

const userItems = (
  state: UserItemsState = initialState,
  action: ActionTypes,
): UserItemsState => {
  switch (action.type) {
    case PATH_NAME_CHANGE:
      // MEMO: ページ遷移したら再度APIを叩かせるようにリセットする
      return initialState;
    case FETCH_USER_ITEMS_SUCCESS:
      return {
        items: state.items.concat(action.payload.items),
        hasNext: action.payload.hasNext,
      };
    case FETCH_USER_PAGE_DATA_SUCCESS:
      return {
        items: state.items.concat(action.payload.items),
        hasNext: action.payload.itemsHasNext,
      };
    default:
      return { ...state };
  }
};

export default userItems;
