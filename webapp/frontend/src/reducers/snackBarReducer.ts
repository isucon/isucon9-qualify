import { ActionTypes } from '../actions/actionTypes';
import { POST_SHIPPED_FAIL } from '../actions/postShippedAction';
import { POST_SHIPPED_DONE_FAIL } from '../actions/postShippedDoneAction';
import { POST_COMPLETE_FAIL } from '../actions/postCompleteAction';
import { SNACK_BAR_CLOSE } from '../actions/snackBarAction';
import { POST_BUMP_FAIL, POST_BUMP_SUCCESS } from '../actions/postBumpAction';
import { SnackBarVariant } from '../components/SnackBar';
import { LOGIN_FAIL } from '../actions/authenticationActions';
import { REGISTER_FAIL } from '../actions/registerAction';
import { PATH_NAME_CHANGE } from '../actions/locationChangeAction';

export interface SnackBarState {
  reason: string;
  available: boolean;
  variant: SnackBarVariant;
}

const initialState: SnackBarState = {
  reason: '',
  available: false,
  variant: 'success',
};

const snackBar = (
  state: SnackBarState = initialState,
  action: ActionTypes,
): SnackBarState => {
  switch (action.type) {
    case LOGIN_FAIL:
    case REGISTER_FAIL:
    case POST_SHIPPED_FAIL:
    case POST_SHIPPED_DONE_FAIL:
    case POST_BUMP_SUCCESS:
    case POST_BUMP_FAIL:
    case POST_COMPLETE_FAIL:
      return {
        reason: action.snackBarMessage,
        available: true,
        variant: action.variant,
      };
    case SNACK_BAR_CLOSE:
    case PATH_NAME_CHANGE:
      return initialState;
    default:
      return { ...state };
  }
};

export default snackBar;
