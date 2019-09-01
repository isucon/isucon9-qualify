import SignUpPageFormComponent from '../components/SignUpFormComponent';
import { connect } from 'react-redux';
import { postRegisterAction } from '../actions/registerAction';
import { RegisterReq } from '../types/appApiTypes';
import { AppState } from '../index';
import { ThunkDispatch } from 'redux-thunk';
import { ActionTypes } from '../actions/actionTypes';

const mapStateToProps = (state: AppState) => ({});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, ActionTypes>,
) => ({
  register: (params: RegisterReq) => {
    dispatch(postRegisterAction(params));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SignUpPageFormComponent);
