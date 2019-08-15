import { AnyAction } from 'redux';
import { TimelineItem } from '../dataObjects/item';
import {
  FETCH_USER_ITEMS_SUCCESS,
  FetchUserItemsSuccessAction,
} from '../actions/fetchUserItemsAction';
import {
  FETCH_USER_PAGE_DATA_SUCCESS,
  FetchUserPageDataSuccessAction,
} from '../actions/fetchUserPageDataAction';
import { LOCATION_CHANGE, LocationChangeAction } from 'connected-react-router';

export interface UserItemsState {
  items: TimelineItem[];
  hasNext: boolean;
}

const initialState: UserItemsState = {
  items: [],
  hasNext: false,
};

type Actions =
  | FetchUserItemsSuccessAction
  | FetchUserPageDataSuccessAction
  | LocationChangeAction
  | AnyAction;

const userItems = (
  state: UserItemsState = initialState,
  action: Actions,
): UserItemsState => {
  switch (action.type) {
    case LOCATION_CHANGE:
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
