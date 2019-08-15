import { connect } from 'react-redux';
import { AppState } from '../index';
import { mockUser } from '../mocks';
import UserPage from '../pages/UserPage';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';
import { fetchTransactionsAction } from '../actions/fetchTransactionsAction';
import { fetchUserItemsAction } from '../actions/fetchUserItemsAction';

const mapStateToProps = (state: AppState) => ({
  loading: true, // TODO state.page.isLoading,
  loggedInUserId: state.authStatus.userId,
  items: state.userItems.items,
  itemsHasNext: state.userItems.hasNext,
  transactions: state.transactions.items,
  transactionsHasNext: state.transactions.hasNext,
  user: mockUser, // TODO
  errorType: state.error.errorType,
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  load: (userId: number) => {
    // TODO
  },
  itemsLoadMore: (
    userId: number,
    itemId: number,
    createdAt: number,
    page: number,
  ) => {
    dispatch(fetchUserItemsAction(userId, itemId, createdAt));
  },
  transactionsLoadMore: (itemId: number, createdAt: number, page: number) => {
    dispatch(fetchTransactionsAction(itemId, createdAt));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(UserPage);
