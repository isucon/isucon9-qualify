import { AnyAction } from 'redux';
import { ItemStatus } from '../dataObjects/item';

type ItemSimple = {
  id: number;
  status: ItemStatus;
  name: string;
  price: number;
  thumbnailUrl: string;
  createdAt: number;
};

export interface TimelineState {
  items: ItemSimple[];
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
