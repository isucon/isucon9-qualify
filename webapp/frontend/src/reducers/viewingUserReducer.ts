import { AnyAction } from 'redux';
import { UserData } from '../dataObjects/user';
import {
  FETCH_USER_PAGE_DATA_SUCCESS,
  FetchUserPageDataSuccessAction,
} from '../actions/fetchUserPageDataAction';

// ユーザページに表示するユーザ情報のstate
export interface ViewingUserState {
  user?: UserData;
}

const initialState: ViewingUserState = {};

type actions = FetchUserPageDataSuccessAction | AnyAction;

const viewingUser = (
  state: ViewingUserState = initialState,
  action: actions,
): ViewingUserState => {
  switch (action.type) {
    case FETCH_USER_PAGE_DATA_SUCCESS:
      return { ...state, user: action.payload.user };
    default:
      return { ...state };
  }
};

export default viewingUser;
