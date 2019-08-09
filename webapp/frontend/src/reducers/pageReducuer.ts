import { AnyAction } from "redux";
import {
  FETCH_ITEM_PAGE_FAIL,
  FETCH_ITEM_PAGE_START,
  FETCH_ITEM_PAGE_SUCCESS
} from "../actions/fetchItemPageAction";
import {
  FETCH_SETTINGS_FAIL,
  FETCH_SETTINGS_START,
  FETCH_SETTINGS_SUCCESS
} from "../actions/settingsAction";

export interface PageState {
  isLoading: boolean;
  isItemPageLoading: boolean;
}

const initialState: PageState = {
  isLoading: true,
  isItemPageLoading: true
};

const page = (
  state: PageState = initialState,
  action: AnyAction
): PageState => {
  switch (action.type) {
    case FETCH_ITEM_PAGE_START:
      return { ...state, isItemPageLoading: true };
    case FETCH_ITEM_PAGE_SUCCESS:
    case FETCH_ITEM_PAGE_FAIL:
      return { ...state, isItemPageLoading: false };
    case FETCH_SETTINGS_START:
      return { ...state, isLoading: true };
    case FETCH_SETTINGS_SUCCESS:
    case FETCH_SETTINGS_FAIL:
      return { ...state, isLoading: false };
    default:
      return { ...state };
  }
};

export default page;
