import { ActionTypes } from '../actions/actionTypes';
import { POST_SHIPPED_FAIL } from '../actions/postShippedAction';
import { POST_SHIPPED_DONE_FAIL } from '../actions/postShippedDoneAction';
import { POST_COMPLETE_FAIL } from '../actions/postCompleteAction';

export interface SnackBarState {
  reason: string;
}

const initialState: SnackBarState = {
  reason: '',
};

const snackBar = (
  state: SnackBarState = initialState,
  action: ActionTypes,
): SnackBarState => {
  switch (action.type) {
    case POST_SHIPPED_FAIL:
    case POST_SHIPPED_DONE_FAIL:
    case POST_COMPLETE_FAIL:
      return { reason: action.snackBarMessage };
    default:
      return initialState;
  }
};

export default snackBar;
