import { connect } from 'react-redux';
import { AppState } from '../index';
import TransactionPage from '../pages/TransactionPage';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';
import { fetchItemAction } from '../actions/fetchItemAction';

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isItemLoading,
  item: state.viewingItem.item,
  auth: {
    userId: state.authStatus.userId || 0,
  },
  errorType: state.error.errorType,
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  load: (itemId: string) => {
    dispatch(fetchItemAction(itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(TransactionPage);
