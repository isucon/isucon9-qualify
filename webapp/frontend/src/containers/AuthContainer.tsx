import {AppState} from "../index";
import {Dispatch} from "redux";
import {connect} from "react-redux";
import {AuthComponent} from "../components/AuthComponent";

const mapStateToProps = (state: AppState) => ({
    isLoggedIn: !!state.authStatus.userId,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({
});

export default connect(mapStateToProps, mapDispatchToProps)(AuthComponent);
