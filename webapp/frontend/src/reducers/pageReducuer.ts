import { AnyAction } from 'redux';
import {
  FETCH_ITEM_FAIL,
  FETCH_ITEM_START,
  FETCH_ITEM_SUCCESS,
  FetchItemFailAction,
  FetchItemStartAction,
  FetchItemSuccessAction,
} from '../actions/fetchItemAction';
import {
  FETCH_SETTINGS_FAIL,
  FETCH_SETTINGS_START,
  FETCH_SETTINGS_SUCCESS,
  FetchSettingsFailAction,
  FetchSettingsStartAction,
  FetchSettingsSuccessAction,
} from '../actions/settingsAction';
import {
  FETCH_TIMELINE_FAIL,
  FETCH_TIMELINE_START,
  FETCH_TIMELINE_SUCCESS,
  FetchTimelineFailAction,
  FetchTimelineStartAction,
  FetchTimelineSuccessAction,
} from '../actions/fetchTimelineAction';
import { LocationChangeAction } from 'connected-react-router';
import {
  Actions as FetchUserPageActions,
  FETCH_USER_PAGE_DATA_FAIL,
  FETCH_USER_PAGE_DATA_START,
  FETCH_USER_PAGE_DATA_SUCCESS,
} from '../actions/fetchUserPageDataAction';
import {
  PATH_NAME_CHANGE,
  PathNameChangeAction,
} from '../actions/locationChangeAction';

type Actions =
  | LocationChangeAction
  | FetchItemStartAction
  | FetchItemSuccessAction
  | FetchItemFailAction
  | FetchTimelineStartAction
  | FetchTimelineSuccessAction
  | FetchTimelineFailAction
  | FetchSettingsStartAction
  | FetchSettingsSuccessAction
  | FetchSettingsFailAction
  | FetchUserPageActions
  | PathNameChangeAction
  | AnyAction;

export interface PageState {
  isLoading: boolean;
  isItemLoading: boolean;
  isTimelineLoading: boolean;
  isUserPageLoading: boolean;
}

const initialState: PageState = {
  isLoading: true,
  isItemLoading: true,
  isTimelineLoading: true,
  isUserPageLoading: true,
};

const pathChangeState: PageState = {
  isLoading: false, // Settings取得時しかtrueにならない
  isItemLoading: true,
  isTimelineLoading: true,
  isUserPageLoading: true,
};

const page = (state: PageState = initialState, action: Actions): PageState => {
  switch (action.type) {
    // Item page
    case FETCH_ITEM_START:
      return { ...state, isItemLoading: true };
    case FETCH_ITEM_SUCCESS:
    case FETCH_ITEM_FAIL:
      return { ...state, isItemLoading: false };
    // Timeline
    case FETCH_TIMELINE_SUCCESS:
    case FETCH_TIMELINE_FAIL:
      return { ...state, isTimelineLoading: false };
    // Settings
    case FETCH_SETTINGS_START:
      return { ...state, isLoading: true };
    case FETCH_SETTINGS_SUCCESS:
    case FETCH_SETTINGS_FAIL:
      return { ...state, isLoading: false };
    // User page
    case FETCH_USER_PAGE_DATA_START:
      return { ...state, isUserPageLoading: true };
    case FETCH_USER_PAGE_DATA_SUCCESS:
    case FETCH_USER_PAGE_DATA_FAIL:
      return { ...state, isUserPageLoading: false };
    // Location change
    case PATH_NAME_CHANGE:
      return pathChangeState;
    default:
      return { ...state };
  }
};

export default page;
