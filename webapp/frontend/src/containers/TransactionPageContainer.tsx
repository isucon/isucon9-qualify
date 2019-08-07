import {connect} from "react-redux";
import {AppState} from "../index";
import TransactionPage from "../pages/TransactionPage";
import {ThunkDispatch} from "redux-thunk";
import {AnyAction} from "redux";

const mapStateToProps = (state: AppState) => ({
    errorType: state.error.errorType,
    loading: false,// TODO state.page.isLoading,
});
const mapDispatchToProps = (dispatch: ThunkDispatch<AppState, undefined, AnyAction>) => ({
});

export default connect(mapStateToProps, mapDispatchToProps)(TransactionPage);
