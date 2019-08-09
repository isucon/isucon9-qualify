import SignUpPageFormComponent from "../components/SignUpFormComponent";
import { connect } from "react-redux";
import { postRegisterAction, State } from "../actions/registerAction";
import { RegisterReq } from "../types/appApiTypes";
import { AppState } from "../index";
import { ThunkDispatch } from "redux-thunk";
import { AnyAction } from "redux";

const mapStateToProps = (state: AppState) => ({
  error: state.formError.error
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<State, undefined, AnyAction>
) => ({
  register: (params: RegisterReq) => {
    dispatch(postRegisterAction(params));
  }
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(SignUpPageFormComponent);
