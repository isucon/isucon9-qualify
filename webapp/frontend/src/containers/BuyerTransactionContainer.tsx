import { AppState } from '../index';
import { AnyAction, Dispatch } from 'redux';
import { connect } from 'react-redux';
import BuyerComponent from '../components/Transaction/BuyerComponent';
import { postShippedAction } from '../actions/postShippedAction';
import { postCompleteAction } from '../actions/postCompleteAction';
import { ThunkDispatch } from 'redux-thunk';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  postShipped: (itemId: number) => {
    dispatch(postShippedAction(itemId));
  },
  postComplete: (itemId: number) => {
    dispatch(postCompleteAction(itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(BuyerComponent);
