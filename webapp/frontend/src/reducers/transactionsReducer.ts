import { AnyAction } from 'redux';
import { TransactionItem } from '../dataObjects/item';
import {
  FETCH_TRANSACTIONS_SUCCESS,
  FetchTransactionsSuccessAction,
} from '../actions/fetchTransactionsAction';

export interface TransactionsState {
  items: TransactionItem[];
  hasNext: boolean;
}

const initialState: TransactionsState = {
  items: [],
  hasNext: false,
};

type Actions = FetchTransactionsSuccessAction | AnyAction;

const transactions = (
  state: TransactionsState = initialState,
  action: Actions,
): TransactionsState => {
  switch (action.type) {
    case FETCH_TRANSACTIONS_SUCCESS:
      return { ...action.payload };
    default:
      return { ...state };
  }
};

export default transactions;
