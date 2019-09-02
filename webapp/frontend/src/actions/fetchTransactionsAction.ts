import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import {
  ItemDetail,
  UserTransactionsReq,
  UserTransactionsRes,
} from '../types/appApiTypes';
import { TransactionItem } from '../dataObjects/item';
import { TransactionStatus } from '../dataObjects/transaction';
import { ShippingStatus } from '../dataObjects/shipping';
import { AppState } from '../index';
import { ErrorActions } from './errorAction';
import { ajaxErrorHandler } from '../actionHelper/ajaxErrorHandler';
import { checkAppResponse } from '../actionHelper/responseChecker';

export const FETCH_TRANSACTIONS_START = 'FETCH_TRANSACTIONS_START';
export const FETCH_TRANSACTIONS_SUCCESS = 'FETCH_TRANSACTIONS_SUCCESS';
export const FETCH_TRANSACTIONS_FAIL = 'FETCH_TRANSACTIONS_FAIL';

export type FetchTransactionActions =
  | FetchTransactionsStartAction
  | FetchTransactionsSuccessAction
  | FetchTransactionsFailAction
  | ErrorActions;
type ThunkResult<R> = ThunkAction<
  R,
  AppState,
  undefined,
  FetchTransactionActions
>;

export function fetchTransactionsAction(
  itemId?: number,
  createdAt?: number,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, FetchTransactionActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(fetchTransactionsStartAction());
      })
      .then(() => {
        return AppClient.get('/users/transactions.json', {
          item_id: itemId,
          created_at: createdAt,
        } as UserTransactionsReq);
      })
      .then(async (response: Response) => {
        await checkAppResponse(response);

        return await response.json();
      })
      .then((body: UserTransactionsRes) => {
        dispatch(
          fetchTransactionsSuccessAction({
            items: body.items.map((item: ItemDetail) => ({
              id: item.id,
              status: item.status,
              transactionEvidenceStatus: item.transaction_evidence_status as TransactionStatus, // MEMO API will return this value for this endpoint
              shippingStatus: item.shipping_status as ShippingStatus, // MEMO API will return this value for this endpoint
              name: item.name,
              thumbnailUrl: item.image_url,
              createdAt: item.created_at,
            })),
            hasNext: body.has_next,
          }),
        );
      })
      .catch(async (err: Error) => {
        dispatch(
          await ajaxErrorHandler<FetchTransactionActions>(
            err,
            fetchTransactionsFailAction,
          ),
        );
      });
  };
}

export interface FetchTransactionsStartAction
  extends Action<typeof FETCH_TRANSACTIONS_START> {}

const fetchTransactionsStartAction = (): FetchTransactionsStartAction => {
  return {
    type: FETCH_TRANSACTIONS_START,
  };
};

export interface FetchTransactionsSuccessAction
  extends Action<typeof FETCH_TRANSACTIONS_SUCCESS> {
  payload: {
    items: TransactionItem[];
    hasNext: boolean;
  };
}

const fetchTransactionsSuccessAction = (payload: {
  items: TransactionItem[];
  hasNext: boolean;
}): FetchTransactionsSuccessAction => {
  return {
    type: FETCH_TRANSACTIONS_SUCCESS,
    payload,
  };
};

export interface FetchTransactionsFailAction
  extends Action<typeof FETCH_TRANSACTIONS_FAIL> {
  message: string;
}

const fetchTransactionsFailAction = (
  message: string,
): FetchTransactionsFailAction => {
  return {
    type: FETCH_TRANSACTIONS_FAIL,
    message,
  };
};
