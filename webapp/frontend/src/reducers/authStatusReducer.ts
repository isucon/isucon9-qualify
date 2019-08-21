import {
  LOGIN_SUCCESS,
  LoginSuccessAction,
} from '../actions/authenticationActions';
import {
  REGISTER_SUCCESS,
  RegisterSuccessAction,
} from '../actions/registerAction';
import {
  FETCH_SETTINGS_FAIL,
  FETCH_SETTINGS_SUCCESS,
  FetchSettingsFailAction,
  FetchSettingsSuccessAction,
} from '../actions/settingsAction';
import { UserData } from '../dataObjects/user';

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

type Actions =
  | LoginSuccessAction
  | RegisterSuccessAction
  | FetchSettingsSuccessAction
  | FetchSettingsFailAction;

const authStatus = (
  state: AuthStatusState = initialState,
  action: Actions,
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
