import { AnyAction } from 'redux';
import { TimelineItem } from '../dataObjects/item';

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

type Actions = AnyAction;

const timeline = (
  state: TimelineState = initialState,
  action: Actions,
): TimelineState => {
  switch (action.type) {
    default:
      return { ...state };
  }
};

export default timeline;
