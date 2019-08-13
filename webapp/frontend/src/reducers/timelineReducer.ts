import { AnyAction } from 'redux';
import { TimelineItem } from '../dataObjects/item';
import {
  FETCH_TIMELINE_SUCCESS,
  FetchTimelineSuccessAction,
} from '../actions/fetchTimelineAction';
import { LOCATION_CHANGE } from 'connected-react-router';

export interface TimelineState {
  items: TimelineItem[];
  hasNext: boolean;
  categoryId?: number;
  categoryName?: string;
}

const initialState: TimelineState = {
  items: [],
  hasNext: false,
};

type Actions = FetchTimelineSuccessAction | AnyAction;

const timeline = (
  state: TimelineState = initialState,
  action: Actions,
): TimelineState => {
  switch (action.type) {
    case LOCATION_CHANGE:
      // MEMO: ページ遷移したらリセットする
      return initialState;
    case FETCH_TIMELINE_SUCCESS:
      const { payload } = action;
      return {
        items: state.items.concat(payload.items),
        hasNext: payload.hasNext,
        categoryId: payload.categoryId,
        categoryName: payload.categoryName,
      };
    default:
      return { ...state };
  }
};

export default timeline;
