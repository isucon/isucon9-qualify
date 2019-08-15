import { AnyAction } from 'redux';
import { TimelineItem } from '../dataObjects/item';
import {
  FETCH_USER_ITEMS_SUCCESS,
  FetchUserItemsSuccessAction,
} from '../actions/fetchUserItemsAction';

export interface UserItemsState {
  items: TimelineItem[];
  hasNext: boolean;
}

const initialState: UserItemsState = {
  items: [],
  hasNext: false,
};

type Actions = FetchUserItemsSuccessAction | AnyAction;

const userItems = (
  state: UserItemsState = initialState,
  action: Actions,
): UserItemsState => {
  switch (action.type) {
    case FETCH_USER_ITEMS_SUCCESS:
      return { ...action.payload };
    default:
      return { ...state };
  }
};

export default userItems;
