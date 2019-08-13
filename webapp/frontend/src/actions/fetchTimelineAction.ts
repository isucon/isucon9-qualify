import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import {
  ItemSimple,
  NewCategoryItemRes,
  NewItemReq,
  NewItemRes,
} from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';
import { TimelineItem } from '../dataObjects/item';
import { NotFoundError } from '../errors/NotFoundError';

export const FETCH_TIMELINE_START = 'FETCH_TIMELINE_START';
export const FETCH_TIMELINE_SUCCESS = 'FETCH_TIMELINE_SUCCESS';
export const FETCH_TIMELINE_FAIL = 'FETCH_TIMELINE_FAIL';

type Actions =
  | FetchTimelineStartAction
  | FetchTimelineSuccessAction
  | FetchTimelineFailAction;
type ThunkResult<R> = ThunkAction<R, void, undefined, Actions>;

export function fetchTimelineAction(
  createdAt?: number,
  itemId?: number,
  categoryId?: number,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, Actions>) => {
    Promise.resolve(() => {
      dispatch(fetchTimelineStartAction());
    })
      .then(() => {
        let getParams: NewItemReq = {
          item_id: itemId,
          created: createdAt,
        };

        if (categoryId) {
          return AppClient.get(`/new_items/${categoryId}.json`, getParams);
        }

        return AppClient.get(`/new_items.json`, getParams);
      })
      .then((response: Response) => {
        if (!response.ok) {
          if (response.status === 404) {
            throw new NotFoundError('Item not found');
          }

          throw new AppResponseError(
            'Request for getting timeline item data was failed',
            response,
          );
        }

        return response.json();
      })
      .then((body: NewItemRes | NewCategoryItemRes) => {
        dispatch(
          fetchTimelineSuccessAction({
            items: body.items.map((item: ItemSimple) => ({
              id: item.id,
              status: item.status,
              name: item.name,
              price: item.price,
              thumbnailUrl:
                'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png', // TODO
              createdAt: item.created_at,
            })),
            hasNext: body.has_next,
            categoryId: body.root_category_id,
            categoryName: body.root_category_name,
          }),
        );
      })
      .catch((err: Error) => {
        dispatch(fetchTimelineFailAction());
      });
  };
}

export interface FetchTimelineStartAction
  extends Action<typeof FETCH_TIMELINE_START> {}

const fetchTimelineStartAction = (): FetchTimelineStartAction => {
  return {
    type: FETCH_TIMELINE_START,
  };
};

export interface FetchTimelineSuccessAction
  extends Action<typeof FETCH_TIMELINE_SUCCESS> {
  payload: {
    items: TimelineItem[];
    hasNext: boolean;
    categoryId?: number;
    categoryName?: string;
  };
}

const fetchTimelineSuccessAction = (payload: {
  items: TimelineItem[];
  hasNext: boolean;
  categoryId?: number;
  categoryName?: string;
}): FetchTimelineSuccessAction => {
  return {
    type: FETCH_TIMELINE_SUCCESS,
    payload,
  };
};

export interface FetchTimelineFailAction
  extends Action<typeof FETCH_TIMELINE_FAIL> {}

const fetchTimelineFailAction = (): FetchTimelineFailAction => {
  return {
    type: FETCH_TIMELINE_FAIL,
  };
};
