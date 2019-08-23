import { TimelineItem } from '../dataObjects/item';
import { FETCH_TIMELINE_SUCCESS } from '../actions/fetchTimelineAction';
import { PATH_NAME_CHANGE } from '../actions/locationChangeAction';
import { ActionTypes } from '../actions/actionTypes';

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

const timeline = (
  state: TimelineState = initialState,
  action: ActionTypes,
): TimelineState => {
  switch (action.type) {
    case PATH_NAME_CHANGE:
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
