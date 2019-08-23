import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import { ErrorRes, GetItemRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';
import { ItemData } from '../dataObjects/item';
import { NotFoundError } from '../errors/NotFoundError';
import { FormErrorState } from '../reducers/formErrorReducer';
import { AppState } from '../index';
import { notFoundError, NotFoundErrorAction } from './errorAction';

export const FETCH_ITEM_START = 'FETCH_ITEM_START';
export const FETCH_ITEM_SUCCESS = 'FETCH_ITEM_SUCCESS';
export const FETCH_ITEM_FAIL = 'FETCH_ITEM_FAIL';

export type FetchItemActions =
  | FetchItemStartAction
  | FetchItemSuccessAction
  | FetchItemFailAction
  | NotFoundErrorAction;

type ThunkResult<R> = ThunkAction<R, AppState, undefined, FetchItemActions>;

export function fetchItemAction(itemId: string): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, FetchItemActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(fetchItemStartAction());
      })
      .then(() => AppClient.get(`/items/${itemId}.json`))
      .then(async (response: Response) => {
        if (!response.ok) {
          if (response.status === 404) {
            throw new NotFoundError('Item not found');
          }

          const errRes: ErrorRes = await response.json();
          throw new AppResponseError(errRes.error, response);
        }

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
      .catch((err: Error) => {
        if (err instanceof NotFoundError) {
          dispatch(notFoundError());
          return;
        }

        dispatch(
          fetchItemFailAction({
            error: err.message,
          }),
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
  payload: FormErrorState;
}

const fetchItemFailAction = (newError: FormErrorState): FetchItemFailAction => {
  return {
    type: FETCH_ITEM_FAIL,
    payload: newError,
  };
};
