import { AppState } from '../index';
import { connect } from 'react-redux';
import { TransactionBuyer } from '../components/TransactionBuyer';
import { postCompleteAction } from '../actions/postCompleteAction';
import { ThunkDispatch } from 'redux-thunk';
import { ActionTypes } from '../actions/actionTypes';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, ActionTypes>,
) => ({
  postComplete: (itemId: number) => {
    dispatch(postCompleteAction(itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(TransactionBuyer);
