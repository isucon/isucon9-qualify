import React from 'react';
import SignUpPageFormComponent from "../components/SignUpFormComponent";
import { connect } from 'react-redux';
import {postRegisterAction} from "../actions/registerAction";
import {RegisterReq} from "../types/appApiTypes";

interface SignUpFormContainerProps {
    register: (params: RegisterReq) => void
    errors: string[],
}

interface SignUpFormContainerState {
}

class SignUpFormContainer extends React.Component<SignUpFormContainerProps, SignUpFormContainerState> {
    render() {
        return (
            <SignUpPageFormComponent
                {...this.props}
            />
        );
    }
}

const mapStateToProps = (state: any) => ({
    errors: state.formError.errorMsg,
});
const mapDispatchToProps = (dispatch: any) => ({
    register: (params: RegisterReq) => {
        dispatch(postRegisterAction(params));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SignUpFormContainer)