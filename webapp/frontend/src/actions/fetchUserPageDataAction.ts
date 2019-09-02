import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import {
  ItemDetail,
  ItemSimple,
  UserItemsRes,
  UserTransactionsRes,
} from '../types/appApiTypes';
import { TimelineItem, TransactionItem } from '../dataObjects/item';
import { UserData } from '../dataObjects/user';
import { ShippingStatus } from '../dataObjects/shipping';
import { TransactionStatus } from '../dataObjects/transaction';
import { AppState } from '../index';
import { ErrorActions } from './errorAction';
import { checkAppResponse } from '../actionHelper/responseChecker';
import { ajaxErrorHandler } from '../actionHelper/ajaxErrorHandler';

export const FETCH_USER_PAGE_DATA_START = 'FETCH_USER_PAGE_DATA_START';
export const FETCH_USER_PAGE_DATA_SUCCESS = 'FETCH_USER_PAGE_DATA_SUCCESS';
export const FETCH_USER_PAGE_DATA_FAIL = 'FETCH_USER_PAGE_DATA_FAIL';

export type FetchUserPageDataActions =
  | FetchUserPageDataStartAction
  | FetchUserPageDataSuccessAction
  | FetchUserPageDataFailAction
  | ErrorActions;
type ThunkResult<R> = ThunkAction<
  R,
  AppState,
  undefined,
  FetchUserPageDataActions
>;

async function fetchUserPageData(
  userId: number,
  isMyPage: boolean,
): Promise<[UserItemsRes, UserTransactionsRes | undefined]> {
  const userDataRes: Response = await AppClient.get(`/users/${userId}.json`);

  await checkAppResponse(userDataRes);

  const userData: UserItemsRes = await userDataRes.json();

  let transactions: UserTransactionsRes | undefined;

  if (isMyPage) {
    const transactionRes: Response = await AppClient.get(
      '/users/transactions.json',
    );

    await checkAppResponse(transactionRes);

    transactions = await transactionRes.json();
  }

  return [userData, transactions];
}

export function fetchUserPageDataAction(
  userId: number,
  isMyPage: boolean,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<AppState, any, FetchUserPageDataActions>) => {
    Promise.resolve()
      .then(() => {
        dispatch(fetchUserPageDataStartAction());
      })
      .then(() => {
        return fetchUserPageData(userId, isMyPage);
      })
      .then((res: [UserItemsRes, UserTransactionsRes | undefined]) => {
        const userPageData = res[0];
        const payload = {
          user: {
            id: userPageData.user.id,
            accountName: userPageData.user.account_name,
            numSellItems: userPageData.user.num_sell_items,
          },
          items: userPageData.items.map((item: ItemSimple) => ({
            id: item.id,
            status: item.status,
            name: item.name,
            price: item.price,
            thumbnailUrl: item.image_url,
            createdAt: item.created_at,
          })),
          itemsHasNext: userPageData.has_next,
        };
        let transactions: {
          transactions: TransactionItem[];
          transactionsHasNext: boolean;
        } = {
          transactions: [],
          transactionsHasNext: false,
        };

        if (isMyPage && res[1] !== undefined) {
          const transactionRes: UserTransactionsRes = res[1];
          transactions = {
            transactions: transactionRes.items.map((item: ItemDetail) => ({
              id: item.id,
              status: item.status,
              transactionEvidenceStatus: item.transaction_evidence_status as TransactionStatus, // MEMO API will return this value for this endpoint
              shippingStatus: item.shipping_status as ShippingStatus, // MEMO API will return this value for this endpoint
              name: item.name,
              thumbnailUrl: item.image_url,
              createdAt: item.created_at,
            })),
            transactionsHasNext: transactionRes.has_next,
          };
        }

        dispatch(
          fetchUserPageDataSuccessAction({ ...payload, ...transactions }),
        );
      })
      .catch(async (err: Error) => {
        dispatch<FetchUserPageDataActions>(
          await ajaxErrorHandler(err, fetchUserPageDataFailAction),
        );
      });
  };
}

export interface FetchUserPageDataStartAction
  extends Action<typeof FETCH_USER_PAGE_DATA_START> {}

const fetchUserPageDataStartAction = (): FetchUserPageDataStartAction => {
  return {
    type: FETCH_USER_PAGE_DATA_START,
  };
};

export interface FetchUserPageDataSuccessAction
  extends Action<typeof FETCH_USER_PAGE_DATA_SUCCESS> {
  payload: {
    user: UserData;
    items: TimelineItem[];
    itemsHasNext: boolean;
    transactions: TransactionItem[];
    transactionsHasNext: boolean;
  };
}

const fetchUserPageDataSuccessAction = (payload: {
  user: UserData;
  items: TimelineItem[];
  itemsHasNext: boolean;
  transactions: TransactionItem[];
  transactionsHasNext: boolean;
}): FetchUserPageDataSuccessAction => {
  return {
    type: FETCH_USER_PAGE_DATA_SUCCESS,
    payload,
  };
};

export interface FetchUserPageDataFailAction
  extends Action<typeof FETCH_USER_PAGE_DATA_FAIL> {
  message: string;
}

const fetchUserPageDataFailAction = (
  message: string,
): FetchUserPageDataFailAction => {
  return { type: FETCH_USER_PAGE_DATA_FAIL, message };
};
