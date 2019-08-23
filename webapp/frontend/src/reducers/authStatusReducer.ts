import { LOGIN_SUCCESS } from '../actions/authenticationActions';
import { REGISTER_SUCCESS } from '../actions/registerAction';
import {
  FETCH_SETTINGS_FAIL,
  FETCH_SETTINGS_SUCCESS,
} from '../actions/settingsAction';
import { UserData } from '../dataObjects/user';
import { ActionTypes } from '../actions/actionTypes';

export interface AuthStatusState {
  userId?: number;
  accountName?: string;
  address?: string;
  numSellItems?: number;
  checked: boolean; // 初回のsettings取得が完了したかどうか
}

const initialState: AuthStatusState = {
  checked: false,
};

const authStatus = (
  state: AuthStatusState = initialState,
  action: ActionTypes,
): AuthStatusState => {
  switch (action.type) {
    case LOGIN_SUCCESS:
    case REGISTER_SUCCESS: {
      return {
        ...state,
        ...action.payload,
      };
    }
    case FETCH_SETTINGS_SUCCESS: {
      const user: UserData | undefined = action.payload.settings.user;
      let userPayload:
        | {
            userId: number;
            accountName: string;
            address?: string;
            numSellItems: number;
          }
        | {} = {};

      if (user) {
        userPayload = {
          userId: user.id,
          accountName: user.accountName,
          address: user.address || undefined,
          numSellItems: user.numSellItems,
        };
      }

      return {
        ...state,
        ...userPayload,
        checked: true,
      };
    }
    case FETCH_SETTINGS_FAIL: {
      return {
        ...state,
        checked: true,
      };
    }
    default:
      return state;
  }
};

export default authStatus;
