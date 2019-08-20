import { connect } from 'react-redux';
import ItemPage from '../pages/ItemPage';
import { fetchItemAction } from '../actions/fetchItemAction';
import { AppState } from '../index';
import { postItemEditAction } from '../actions/postItemEditAction';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isItemLoading,
  item: state.viewingItem.item,
  errorType: state.error.errorType,
  formError: 'TODO',
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  load: (itemId: string) => {
    dispatch(fetchItemAction(itemId));
  },
  onClickEdit: (itemId: number, price: number) => {
    dispatch(postItemEditAction(itemId, price));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemPage);
