import SignInPageFormComponent from "../components/SignInFormComponent";
import { connect } from "react-redux";
import { postLoginAction } from "../actions/authenticationActions";
import { AppState } from "../index";
import { ThunkDispatch } from "redux-thunk";
import { AnyAction } from "redux";

const mapStateToProps = (state: AppState) => ({
  error: state.formError.error
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>
) => ({
  onSubmit: (accountName: string, password: string) => {
    dispatch(postLoginAction(accountName, password));
  }
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(SignInPageFormComponent);
