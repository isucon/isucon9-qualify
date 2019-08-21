import { AppState } from '../index';
import { AnyAction } from 'redux';
import { connect } from 'react-redux';
import BuyerComponent from '../components/Transaction/BuyerComponent';
import { postCompleteAction } from '../actions/postCompleteAction';
import { ThunkDispatch } from 'redux-thunk';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  postComplete: (itemId: number) => {
    dispatch(postCompleteAction(itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(BuyerComponent);
