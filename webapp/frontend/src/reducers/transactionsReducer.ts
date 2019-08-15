import { AnyAction } from 'redux';
import { TransactionItem } from '../dataObjects/item';

export interface TransactionsState {
  items: TransactionItem[];
  hasNext: boolean;
}

const initialState: TransactionsState = {
  items: [],
  hasNext: false,
};

type Actions = AnyAction;

const transactions = (
  state: TransactionsState = initialState,
  action: Actions,
): TransactionsState => {
  switch (action.type) {
    default:
      return { ...state };
  }
};

export default transactions;
