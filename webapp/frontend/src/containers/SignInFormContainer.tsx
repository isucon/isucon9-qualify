import SignInPageFormComponent from "../components/SignInFormComponent";
import { connect } from 'react-redux';
import { postLoginAction } from "../actions/authenticationActions";

const mapStateToProps = (state: any) => ({
    error: state.formError.error,
});
const mapDispatchToProps = (dispatch: any) => ({
    onSubmit: (accountName: string, password: string) => {
        dispatch(postLoginAction(accountName, password));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SignInPageFormComponent)