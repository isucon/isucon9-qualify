import { AppState } from '../index';
import { AnyAction } from 'redux';
import { connect } from 'react-redux';
import SellerComponent from '../components/Transaction/SellerComponent';
import { postShippedDoneAction } from '../actions/postShippedDoneAction';
import { ThunkDispatch } from 'redux-thunk';
import { postShippedAction } from '../actions/postShippedAction';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  postShipped: (itemId: number) => {
    dispatch(postShippedAction(itemId));
  },
  postShippedDone: (itemId: number) => {
    dispatch(postShippedDoneAction(itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SellerComponent);
