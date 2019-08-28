import { TransactionItem } from '../dataObjects/item';
import { FETCH_TRANSACTIONS_SUCCESS } from '../actions/fetchTransactionsAction';
import { FETCH_USER_PAGE_DATA_SUCCESS } from '../actions/fetchUserPageDataAction';
import { ActionTypes } from '../actions/actionTypes';
import { PATH_NAME_CHANGE } from '../actions/locationChangeAction';

export interface TransactionsState {
  items: TransactionItem[];
  hasNext: boolean;
}

const initialState: TransactionsState = {
  items: [],
  hasNext: false,
};

const transactions = (
  state: TransactionsState = initialState,
  action: ActionTypes,
): TransactionsState => {
  switch (action.type) {
    case PATH_NAME_CHANGE:
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
