import {AppState} from "../index";
import {Dispatch} from "redux";
import {connect} from "react-redux";
import {NonAuthComponent} from "../components/NonAuthComponent";

const mapStateToProps = (state: AppState) => ({
    isLoggedIn: !!state.authStatus.userId,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({
});

export default connect(mapStateToProps, mapDispatchToProps)(NonAuthComponent);
