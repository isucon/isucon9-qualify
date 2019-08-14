import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action, AnyAction } from 'redux';
import { GetItemRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';
import { ItemData } from '../dataObjects/item';
import { NotFoundError } from '../errors/NotFoundError';

export const FETCH_ITEM_START = 'FETCH_ITEM_START';
export const FETCH_ITEM_SUCCESS = 'FETCH_ITEM_SUCCESS';
export const FETCH_ITEM_FAIL = 'FETCH_ITEM_FAIL';

type ThunkResult<R> = ThunkAction<R, void, undefined, AnyAction>;

export function fetchItemAction(itemId: string): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, AnyAction>) => {
    Promise.resolve(() => {
      dispatch(fetchItemStartAction());
    })
      .then(() => AppClient.get(`/items/${itemId}.json`))
      .then((response: Response) => {
        if (!response.ok) {
          if (response.status === 404) {
            throw new NotFoundError('Item not found');
          }

          throw new AppResponseError(
            'Request for getting item data was failed',
            response,
          );
        }

        return response.json();
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
            thumbnailUrl:
              'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png', // TODO
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
        dispatch(fetchItemFailAction());
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

export interface FetchItemFailAction extends Action<typeof FETCH_ITEM_FAIL> {}

const fetchItemFailAction = (): FetchItemFailAction => {
  return {
    type: FETCH_ITEM_FAIL,
  };
};
