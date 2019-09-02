import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { GetItemRes } from '../types/appApiTypes';
import { ItemData } from '../dataObjects/item';
import { AppState } from '../index';
import { ErrorActions } from './errorAction';
import { checkAppResponse } from '../actionHelper/responseChecker';
import { ajaxErrorHandler } from '../actionHelper/ajaxErrorHandler';

export const FETCH_ITEM_START = 'FETCH_ITEM_START';
export const FETCH_ITEM_SUCCESS = 'FETCH_ITEM_SUCCESS';
export const FETCH_ITEM_FAIL = 'FETCH_ITEM_FAIL';

export type FetchItemActions =
  | FetchItemStartAction
  | FetchItemSuccessAction
  | FetchItemFailAction
  | ErrorActions;

type ThunkResult<R> = ThunkAction<R, AppState, undefined, FetchItemActions>;

export function fetchItemAction(itemId: string): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, FetchItemActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(fetchItemStartAction());
      })
      .then(() => AppClient.get(`/items/${itemId}.json`))
      .then(async (response: Response) => {
        await checkAppResponse(response);

        return await response.json();
      })
      .then((body: GetItemRes) => {
        dispatch(
          fetchItemSuccessAction({
            id: body.id,
            status: body.status,
            sellerId: body.seller_id,
            seller: {
              id: body.seller.id,
              accountName: body.seller.account_name,
              numSellItems: body.seller.num_sell_items,
            },
            buyerId: body.buyer_id,
            buyer: body.buyer,
            name: body.name,
            price: body.price,
            thumbnailUrl: body.image_url,
            description: body.description,
            category: {
              id: body.category.id,
              parentId: body.category.parent_id,
              categoryName: body.category.category_name,
              parentCategoryName: body.category.parent_category_name,
            },
            transactionEvidenceId: body.transaction_evidence_id,
            transactionEvidenceStatus: body.transaction_evidence_status,
            shippingStatus: body.shipping_status,
            createdAt: body.created_at,
          }),
        );
      })
      .catch(async (err: Error) => {
        return dispatch(
          await ajaxErrorHandler<FetchItemActions>(err, fetchItemFailAction),
        );
      });
  };
}

export interface FetchItemStartAction extends Action<typeof FETCH_ITEM_START> {}

const fetchItemStartAction = (): FetchItemStartAction => {
  return {
    type: FETCH_ITEM_START,
  };
};

export interface FetchItemSuccessAction
  extends Action<typeof FETCH_ITEM_SUCCESS> {
  payload: {
    item: ItemData;
  };
}

const fetchItemSuccessAction = (item: ItemData): FetchItemSuccessAction => {
  return {
    type: FETCH_ITEM_SUCCESS,
    payload: {
      item,
    },
  };
};

export interface FetchItemFailAction extends Action<typeof FETCH_ITEM_FAIL> {
  message: string;
}

const fetchItemFailAction = (message: string): FetchItemFailAction => {
  return {
    type: FETCH_ITEM_FAIL,
    message,
  };
};
