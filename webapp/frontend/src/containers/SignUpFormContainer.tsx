import SignUpPageFormComponent from "../components/SignUpFormComponent";
import { connect } from 'react-redux';
import {postRegisterAction} from "../actions/registerAction";
import {RegisterReq} from "../types/appApiTypes";

const mapStateToProps = (state: any) => ({
    errors: state.formError.errorMsg,
});
const mapDispatchToProps = (dispatch: any) => ({
    register: (params: RegisterReq) => {
        dispatch(postRegisterAction(params));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SignUpPageFormComponent)