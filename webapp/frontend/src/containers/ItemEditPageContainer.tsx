import { connect } from 'react-redux';
import { fetchItemAction } from '../actions/fetchItemAction';
import { AppState } from '../index';
import { postItemEditAction } from '../actions/postItemEditAction';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';
import ItemEditPage from '../pages/ItemEditPage';

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isItemLoading,
  item: state.viewingItem.item,
  errorType: state.error.errorType,
  formError: state.formError.itemEditFormError,
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
)(ItemEditPage);
