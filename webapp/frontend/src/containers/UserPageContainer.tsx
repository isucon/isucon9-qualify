import { connect } from 'react-redux';
import { AppState } from '../index';
import { mockItems, mockUser } from '../mocks';
import UserPage from '../pages/UserPage';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';
import { fetchTransactionsAction } from '../actions/fetchTransactionsAction';

const mapStateToProps = (state: AppState) => ({
  loading: true, // TODO state.page.isLoading,
  loggedInUserId: state.authStatus.userId,
  items: mockItems, // TODO
  itemsHasNext: false, // TODO
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
    // TODO
  },
  transactionsLoadMore: (itemId: number, createdAt: number, page: number) => {
    dispatch(fetchTransactionsAction(itemId, createdAt));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(UserPage);
