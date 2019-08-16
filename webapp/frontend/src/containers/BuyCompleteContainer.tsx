import { connect } from 'react-redux';
import BuyCompletePage from '../pages/BuyComplete';
import { Dispatch } from 'redux';
import { push } from 'connected-react-router';
import { routes } from '../routes/Route';

const mapStateToProps = (state: any) => ({
  itemId: state.viewingItem.item.id || 0,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({
  onClickTransaction: (itemId: number) => {
    dispatch(push(routes.transaction.getPath(itemId)));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(BuyCompletePage);
