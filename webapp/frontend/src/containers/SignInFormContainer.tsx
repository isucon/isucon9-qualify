import React from 'react';
import SignInPageFormComponent from "../components/SignInFormComponent";
import { connect } from 'react-redux';
import { postLoginAction } from "../actions/authenticationActions";

interface SignInFormContainerProps {
    onSubmit: (accountName: string, password: string) => void
    errors: string[],
}

interface SignInFormContainerState {
}

class SignInFormContainer extends React.Component<SignInFormContainerProps, SignInFormContainerState> {
    render() {
        return (
            <SignInPageFormComponent
                {...this.props}
            />
        );
    }
}

const mapStateToProps = (state: any) => ({
    errors: state.formError.errorMsg,
});
const mapDispatchToProps = (dispatch: any) => ({
    onSubmit: (accountName: string, password: string) => {
        dispatch(postLoginAction(accountName, password));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SignInFormContainer)