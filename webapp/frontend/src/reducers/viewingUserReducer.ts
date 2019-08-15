import { AnyAction } from 'redux';
import { UserData } from '../dataObjects/user';

// ユーザページに表示するユーザ情報のstate
export interface ViewingUserState {
  user?: UserData;
}

const initialState: ViewingUserState = {};

type actions = AnyAction;

const viewingUser = (
  state: ViewingUserState = initialState,
  action: actions,
): ViewingUserState => {
  switch (action.type) {
    default:
      return { ...state };
  }
};

export default viewingUser;
