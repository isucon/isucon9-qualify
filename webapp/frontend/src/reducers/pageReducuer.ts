import { AnyAction } from 'redux';
import {
  FETCH_ITEM_PAGE_FAIL,
  FETCH_ITEM_PAGE_START,
  FETCH_ITEM_PAGE_SUCCESS,
  FetchItemPageFailAction,
  FetchItemPageStartAction,
  FetchItemPageSuccessAction,
} from '../actions/fetchItemPageAction';
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
import { LOCATION_CHANGE, LocationChangeAction } from 'connected-react-router';
import { routes } from '../routes/Route';

type Actions =
  | LocationChangeAction
  | FetchItemPageStartAction
  | FetchItemPageSuccessAction
  | FetchItemPageFailAction
  | FetchTimelineStartAction
  | FetchTimelineSuccessAction
  | FetchTimelineFailAction
  | FetchSettingsStartAction
  | FetchSettingsSuccessAction
  | FetchSettingsFailAction
  | AnyAction;

export interface PageState {
  isLoading: boolean;
  isItemPageLoading: boolean;
  isTimelineLoading: boolean;
}

const initialState: PageState = {
  isLoading: true,
  isItemPageLoading: true,
  isTimelineLoading: true,
};

const page = (state: PageState = initialState, action: Actions): PageState => {
  switch (action.type) {
    // Item page
    case FETCH_ITEM_PAGE_START:
      return { ...state, isItemPageLoading: true };
    case FETCH_ITEM_PAGE_SUCCESS:
    case FETCH_ITEM_PAGE_FAIL:
      return { ...state, isItemPageLoading: false };
    // Timeline
    case FETCH_TIMELINE_START:
      return { ...state, isTimelineLoading: true };
    case FETCH_TIMELINE_SUCCESS:
    case FETCH_TIMELINE_FAIL:
      return { ...state, isTimelineLoading: false };
    // Settings
    case FETCH_SETTINGS_START:
      return { ...state, isLoading: true };
    case FETCH_SETTINGS_SUCCESS:
    case FETCH_SETTINGS_FAIL:
      return { ...state, isLoading: false };
    case LOCATION_CHANGE:
      const {
        payload: {
          location: { pathname },
        },
      } = action as LocationChangeAction; // TODO なんでasつけないと動かないん？

      switch (pathname) {
        case routes.timeline.path:
          // TODO カテゴリ新着のチェックもここに入る
          return { ...state, isTimelineLoading: true };
        default:
          return { ...state };
      }
    default:
      return { ...state };
  }
};

export default page;
