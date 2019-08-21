import { connect } from 'react-redux';
import ItemPage from '../pages/ItemPage';
import { fetchItemAction } from '../actions/fetchItemAction';
import { AppState } from '../index';
import { push } from 'connected-react-router';
import { routes } from '../routes/Route';
import { postBumpAction } from '../actions/postBumpAction';

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isItemLoading,
  item: state.viewingItem.item,
  viewer: {
    userId: state.authStatus.userId || 0,
  },
  errorType: state.error.errorType,
});
const mapDispatchToProps = (dispatch: any) => ({
  load: (itemId: string) => {
    dispatch(fetchItemAction(itemId));
  },
  onClickBuy: (itemId: number) => {
    dispatch(push(routes.buy.getPath(itemId)));
  },
  onClickItemEdit: (itemId: number) => {
    dispatch(push(routes.itemEdit.getPath(itemId)));
  },
  onClickBump: (itemId: number) => {
    dispatch(postBumpAction(itemId));
  },
  onClickTransaction: (itemId: number) => {
    dispatch(push(routes.transaction.getPath(itemId)));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemPage);
