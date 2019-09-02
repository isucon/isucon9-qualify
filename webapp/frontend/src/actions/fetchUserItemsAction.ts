import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { ItemSimple, UserItemsReq, UserItemsRes } from '../types/appApiTypes';
import { TimelineItem } from '../dataObjects/item';
import { AppState } from '../index';
import { ErrorActions } from './errorAction';
import { checkAppResponse } from '../actionHelper/responseChecker';
import { ajaxErrorHandler } from '../actionHelper/ajaxErrorHandler';

export const FETCH_USER_ITEMS_START = 'FETCH_USER_ITEMS_START';
export const FETCH_USER_ITEMS_SUCCESS = 'FETCH_USER_ITEMS_SUCCESS';
export const FETCH_USER_ITEMS_FAIL = 'FETCH_USER_ITEMS_FAIL';

export type FetchUserItemsActions =
  | FetchUserItemsStartAction
  | FetchUserItemsSuccessAction
  | FetchUserItemsFailAction
  | ErrorActions;
type ThunkResult<R> = ThunkAction<
  R,
  AppState,
  undefined,
  FetchUserItemsActions
>;

export function fetchUserItemsAction(
  userId: number,
  itemId?: number,
  createdAt?: number,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, FetchUserItemsActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(fetchUserItemsStartAction());
      })
      .then(() => {
        return AppClient.get(`/users/${userId}.json`, {
          item_id: itemId,
          created_at: createdAt,
        } as UserItemsReq);
      })
      .then(async (response: Response) => {
        await checkAppResponse(response);

        return await response.json();
      })
      .then((body: UserItemsRes) => {
        dispatch(
          fetchUserItemsSuccessAction({
            items: body.items.map((item: ItemSimple) => ({
              id: item.id,
              status: item.status,
              name: item.name,
              price: item.price,
              thumbnailUrl: item.image_url,
              createdAt: item.created_at,
            })),
            hasNext: body.has_next,
          }),
        );
      })
      .catch(async (err: Error) => {
        dispatch(
          await ajaxErrorHandler<FetchUserItemsActions>(
            err,
            fetchUserItemsFailAction,
          ),
        );
      });
  };
}

export interface FetchUserItemsStartAction
  extends Action<typeof FETCH_USER_ITEMS_START> {}

const fetchUserItemsStartAction = (): FetchUserItemsStartAction => {
  return {
    type: FETCH_USER_ITEMS_START,
  };
};

export interface FetchUserItemsSuccessAction
  extends Action<typeof FETCH_USER_ITEMS_SUCCESS> {
  payload: {
    items: TimelineItem[];
    hasNext: boolean;
  };
}

const fetchUserItemsSuccessAction = (payload: {
  items: TimelineItem[];
  hasNext: boolean;
}): FetchUserItemsSuccessAction => {
  return {
    type: FETCH_USER_ITEMS_SUCCESS,
    payload,
  };
};

export interface FetchUserItemsFailAction
  extends Action<typeof FETCH_USER_ITEMS_FAIL> {
  message: string;
}

const fetchUserItemsFailAction = (
  message: string,
): FetchUserItemsFailAction => {
  return { type: FETCH_USER_ITEMS_FAIL, message };
};
