import React from 'react';
import SignInPageFormComponent from "../components/SignInFormComponent";
import { connect } from 'react-redux';
import { postLoginAction } from "../actions/authenticationActions";

interface SignInFormContainerProps {
    onSubmit: (userId: string, password: string) => void
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
    onSubmit: (userId: string, password: string) => {
        dispatch(postLoginAction(userId, password));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SignInFormContainer)