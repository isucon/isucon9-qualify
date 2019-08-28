import { AppState } from '../index';
import { connect } from 'react-redux';
import { TransactionSeller } from '../components/TransactionSeller';
import { postShippedDoneAction } from '../actions/postShippedDoneAction';
import { ThunkDispatch } from 'redux-thunk';
import { postShippedAction } from '../actions/postShippedAction';
import { ActionTypes } from '../actions/actionTypes';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, ActionTypes>,
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
)(TransactionSeller);
