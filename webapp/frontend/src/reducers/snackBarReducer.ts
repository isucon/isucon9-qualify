import { ActionTypes } from '../actions/actionTypes';
import { POST_SHIPPED_FAIL } from '../actions/postShippedAction';
import { POST_SHIPPED_DONE_FAIL } from '../actions/postShippedDoneAction';
import { POST_COMPLETE_FAIL } from '../actions/postCompleteAction';
import { SNACK_BAR_CLOSE } from '../actions/snackBarAction';

export interface SnackBarState {
  reason: string;
  available: boolean;
}

const initialState: SnackBarState = {
  reason: '',
  available: false,
};

const snackBar = (
  state: SnackBarState = initialState,
  action: ActionTypes,
): SnackBarState => {
  switch (action.type) {
    case POST_SHIPPED_FAIL:
    case POST_SHIPPED_DONE_FAIL:
    case POST_COMPLETE_FAIL:
      return {
        reason: action.snackBarMessage,
        available: true,
      };
    case SNACK_BAR_CLOSE:
      return {
        reason: '',
        available: false,
      };
    default:
      return { ...state };
  }
};

export default snackBar;
