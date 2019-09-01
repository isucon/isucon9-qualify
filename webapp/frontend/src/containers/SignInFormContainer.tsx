import SignInPageFormComponent from '../components/SignInFormComponent';
import { connect } from 'react-redux';
import { postLoginAction } from '../actions/authenticationActions';
import { AppState } from '../index';
import { ThunkDispatch } from 'redux-thunk';
import { ActionTypes } from '../actions/actionTypes';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, ActionTypes>,
) => ({
  onSubmit: (accountName: string, password: string) => {
    dispatch(postLoginAction(accountName, password));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SignInPageFormComponent);
