import { connect } from 'react-redux';
import { fetchItemAction } from '../actions/fetchItemAction';
import { AppState } from '../index';
import ItemBuyPage from '../pages/ItemBuyPage';

const mapStateToProps = (state: AppState) => ({
  loading: !state.viewingItem.item, // 商品がstateにない場合はloadingにする
  item: state.viewingItem.item,
  errorType: state.error.errorType,
});
const mapDispatchToProps = (dispatch: any) => ({
  load: (itemId: string) => {
    dispatch(fetchItemAction(itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemBuyPage);
