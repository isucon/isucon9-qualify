import {
  INTERNAL_SERVER_ERROR,
  InternalServerErrorAction,
  NOT_FOUND_ERROR,
  NotFoundErrorAction,
} from '../actions/errorAction';
import {
  FETCH_ITEM_FAIL,
  FetchItemFailAction,
} from '../actions/fetchItemAction';
import {
  FETCH_SETTINGS_FAIL,
  FetchSettingsFailAction,
} from '../actions/settingsAction';
import {
  POST_SHIPPED_FAIL,
  PostShippedFailAction,
} from '../actions/postShippedAction';
import {
  POST_SHIPPED_DONE_FAIL,
  PostShippedDoneFailAction,
} from '../actions/postShippedDoneAction';
import {
  POST_COMPLETE_FAIL,
  PostCompleteFailAction,
} from '../actions/postCompleteAction';
import {
  FETCH_TRANSACTIONS_FAIL,
  FetchTransactionsFailAction,
} from '../actions/fetchTransactionsAction';
import {
  FETCH_USER_ITEMS_FAIL,
  FetchUserItemsFailAction,
} from '../actions/fetchUserItemsAction';
import {
  FETCH_USER_PAGE_DATA_FAIL,
  FetchUserPageDataFailAction,
} from '../actions/fetchUserPageDataAction';
import {
  FETCH_TIMELINE_FAIL,
  FetchTimelineFailAction,
} from '../actions/fetchTimelineAction';

export const NoError = 'NO_ERROR';
export const NotFoundError = 'NOT_FOUND';
export const InternalServerError = 'INTERNAL_SERVER_ERROR';
export type ErrorType =
  | typeof NoError
  | typeof NotFoundError
  | typeof InternalServerError;

export interface ErrorState {
  errorType: ErrorType;
  errorCode?: number;
}

const initialState: ErrorState = {
  errorType: NoError,
};

type errorActions =
  | NotFoundErrorAction
  | InternalServerErrorAction
  | FetchItemFailAction
  | FetchTimelineFailAction
  | FetchTransactionsFailAction
  | FetchUserItemsFailAction
  | FetchUserPageDataFailAction
  | FetchSettingsFailAction
  | PostShippedFailAction
  | PostShippedDoneFailAction
  | PostCompleteFailAction;

const error = (
  state: ErrorState = initialState,
  action: errorActions,
): ErrorState => {
  switch (action.type) {
    case NOT_FOUND_ERROR:
      return { errorType: NotFoundError, errorCode: 404 };
    case INTERNAL_SERVER_ERROR:
    case FETCH_ITEM_FAIL:
    case FETCH_TIMELINE_FAIL:
    case FETCH_TRANSACTIONS_FAIL:
    case FETCH_USER_ITEMS_FAIL:
    case FETCH_USER_PAGE_DATA_FAIL:
    case FETCH_SETTINGS_FAIL:
    case POST_SHIPPED_FAIL:
    case POST_SHIPPED_DONE_FAIL:
    case POST_COMPLETE_FAIL:
      return { errorType: InternalServerError, errorCode: 500 };
    default:
      return { errorType: NoError };
  }
};

export default error;
