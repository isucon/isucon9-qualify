import { AnyAction } from 'redux';
import { TransactionItem } from '../dataObjects/item';
import {
  FETCH_TRANSACTIONS_SUCCESS,
  FetchTransactionsSuccessAction,
} from '../actions/fetchTransactionsAction';
import {
  FETCH_USER_PAGE_DATA_SUCCESS,
  FetchUserPageDataSuccessAction,
} from '../actions/fetchUserPageDataAction';
import { LOCATION_CHANGE, LocationChangeAction } from 'connected-react-router';

export interface TransactionsState {
  items: TransactionItem[];
  hasNext: boolean;
}

const initialState: TransactionsState = {
  items: [],
  hasNext: false,
};

type Actions =
  | FetchTransactionsSuccessAction
  | LocationChangeAction
  | FetchUserPageDataSuccessAction
  | AnyAction;

const transactions = (
  state: TransactionsState = initialState,
  action: Actions,
): TransactionsState => {
  switch (action.type) {
    case LOCATION_CHANGE:
      // MEMO: ページ遷移したら再度APIを叩かせるようにリセットする
      return initialState;
    case FETCH_TRANSACTIONS_SUCCESS:
      return {
        items: state.items.concat(action.payload.items),
        hasNext: action.payload.hasNext,
      };
    case FETCH_USER_PAGE_DATA_SUCCESS:
      return {
        items: state.items.concat(action.payload.transactions),
        hasNext: action.payload.transactionsHasNext,
      };
    default:
      return { ...state };
  }
};

export default transactions;
