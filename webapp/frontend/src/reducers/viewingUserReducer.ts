import { UserData } from '../dataObjects/user';
import { FETCH_USER_PAGE_DATA_SUCCESS } from '../actions/fetchUserPageDataAction';
import { ActionTypes } from '../actions/actionTypes';
import { PATH_NAME_CHANGE } from '../actions/locationChangeAction';

// ユーザページに表示するユーザ情報のstate
export interface ViewingUserState {
  user?: UserData;
}

const initialState: ViewingUserState = {};

const viewingUser = (
  state: ViewingUserState = initialState,
  action: ActionTypes,
): ViewingUserState => {
  switch (action.type) {
    case PATH_NAME_CHANGE:
      return initialState;
    case FETCH_USER_PAGE_DATA_SUCCESS:
      return { ...state, user: action.payload.user };
    default:
      return { ...state };
  }
};

export default viewingUser;
