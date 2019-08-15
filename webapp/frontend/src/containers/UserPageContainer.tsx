import { connect } from 'react-redux';
import { AppState } from '../index';
import UserPage from '../pages/UserPage';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';
import { fetchTransactionsAction } from '../actions/fetchTransactionsAction';
import { fetchUserItemsAction } from '../actions/fetchUserItemsAction';
import { fetchUserPageDataAction } from '../actions/fetchUserPageDataAction';

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isUserPageLoading,
  loggedInUserId: state.authStatus.userId,
  items: state.userItems.items,
  itemsHasNext: state.userItems.hasNext,
  transactions: state.transactions.items,
  transactionsHasNext: state.transactions.hasNext,
  user: state.viewingUser.user,
  errorType: state.error.errorType,
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  load: (userId: number, isMyPage: boolean) => {
    dispatch(fetchUserPageDataAction(userId, isMyPage));
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
