import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import {
  ItemSimple,
  NewCategoryItemRes,
  NewItemReq,
  NewItemRes,
} from '../types/appApiTypes';
import { TimelineItem } from '../dataObjects/item';
import { AppState } from '../index';
import { ErrorActions } from './errorAction';
import { checkAppResponse } from '../actionHelper/responseChecker';
import { ajaxErrorHandler } from '../actionHelper/ajaxErrorHandler';

export const FETCH_TIMELINE_START = 'FETCH_TIMELINE_START';
export const FETCH_TIMELINE_SUCCESS = 'FETCH_TIMELINE_SUCCESS';
export const FETCH_TIMELINE_FAIL = 'FETCH_TIMELINE_FAIL';

export type FetchTimelineActions =
  | FetchTimelineStartAction
  | FetchTimelineSuccessAction
  | FetchTimelineFailAction
  | ErrorActions;
type ThunkResult<R> = ThunkAction<R, AppState, undefined, FetchTimelineActions>;

export function fetchTimelineAction(
  createdAt?: number,
  itemId?: number,
  categoryId?: number,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, FetchTimelineActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(fetchTimelineStartAction());
      })
      .then(() => {
        let getParams: NewItemReq = {
          item_id: itemId,
          created_at: createdAt,
        };

        if (categoryId) {
          return AppClient.get(`/new_items/${categoryId}.json`, getParams);
        }

        return AppClient.get(`/new_items.json`, getParams);
      })
      .then(async (response: Response) => {
        await checkAppResponse(response);

        return await response.json();
      })
      .then((body: NewItemRes | NewCategoryItemRes) => {
        dispatch(
          fetchTimelineSuccessAction({
            items: body.items.map((item: ItemSimple) => ({
              id: item.id,
              status: item.status,
              name: item.name,
              price: item.price,
              thumbnailUrl: item.image_url,
              createdAt: item.created_at,
            })),
            hasNext: body.has_next,
            categoryId: body.root_category_id,
            categoryName: body.root_category_name,
          }),
        );
      })
      .catch(async (err: Error) => {
        dispatch(
          await ajaxErrorHandler<FetchTimelineActions>(
            err,
            fetchTimelineFailAction,
          ),
        );
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
  extends Action<typeof FETCH_TIMELINE_FAIL> {
  message: string;
}

const fetchTimelineFailAction = (message: string): FetchTimelineFailAction => {
  return {
    type: FETCH_TIMELINE_FAIL,
    message,
  };
};
