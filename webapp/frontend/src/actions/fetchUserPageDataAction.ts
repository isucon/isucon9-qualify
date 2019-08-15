import AppClient from '../httpClients/appClient';
import { ThunkAction, ThunkDispatch } from 'redux-thunk';
import { Action } from 'redux';
import {
  ItemDetail,
  ItemSimple,
  UserItemsRes,
  UserTransactionsRes,
} from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';
import { TimelineItem, TransactionItem } from '../dataObjects/item';
import { UserData } from '../dataObjects/user';
import { ShippingStatus } from '../dataObjects/shipping';
import { TransactionStatus } from '../dataObjects/transaction';
import { FormErrorState } from '../reducers/formErrorReducer';

export const FETCH_USER_PAGE_DATA_START = 'FETCH_USER_PAGE_DATA_START';
export const FETCH_USER_PAGE_DATA_SUCCESS = 'FETCH_USER_PAGE_DATA_SUCCESS';
export const FETCH_USER_PAGE_DATA_FAIL = 'FETCH_USER_PAGE_DATA_FAIL';

export type Actions =
  | FetchUserPageDataStartAction
  | FetchUserPageDataSuccessAction
  | FetchUserPageDataFailAction;
type ThunkResult<R> = ThunkAction<R, void, undefined, Actions>;

async function fetchUserPaggeData(
  userId: number,
  isMyPage: boolean,
): Promise<[UserItemsRes, UserTransactionsRes | undefined]> {
  const userDataRes: Response = await AppClient.get(`/users/${userId}.json`);

  if (!userDataRes.ok) {
    throw new AppResponseError(
      `Fetching data from /users/${userId} was failed`,
      userDataRes,
    );
  }

  const userData: UserItemsRes = await userDataRes.json();

  let transactions: UserTransactionsRes | undefined;

  if (isMyPage) {
    const transactionRes: Response = await AppClient.get(
      '/users/transactions.json',
    );

    if (!transactionRes.ok) {
      throw new AppResponseError(
        `Fetching data from /users/transactions.json was failed`,
        transactionRes,
      );
    }

    transactions = await transactionRes.json();
  }

  return [userData, transactions];
}

export function fetchUserPageDataAction(
  userId: number,
  isMyPage: boolean,
): ThunkResult<void> {
  return (dispatch: ThunkDispatch<any, any, Actions>) => {
    Promise.resolve(() => {
      dispatch(fetchUserPageDataStartAction());
    })
      .then(() => {
        return fetchUserPaggeData(userId, isMyPage);
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
            thumbnailUrl:
              'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png', // TODO
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
              thumbnailUrl:
                'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png', // TODO
              createdAt: item.created_at,
            })),
            transactionsHasNext: transactionRes.has_next,
          };
        }

        dispatch(
          fetchUserPageDataSuccessAction({ ...payload, ...transactions }),
        );
      })
      .catch((err: Error) => {
        dispatch(
          fetchUserPageDataFailAction({
            error: err.message,
          }),
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
  payload: FormErrorState;
}

const fetchUserPageDataFailAction = (
  newError: FormErrorState,
): FetchUserPageDataFailAction => {
  return {
    type: FETCH_USER_PAGE_DATA_FAIL,
    payload: newError,
  };
};
