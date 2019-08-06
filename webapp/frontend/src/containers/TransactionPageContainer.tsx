import {connect} from "react-redux";
import {AppState} from "../index";
import TransactionPage from "../pages/TransactionPage";

const mapStateToProps = (state: AppState) => ({
    errorType: state.error.errorType,
    loading: false,// TODO state.page.isLoading,
});
const mapDispatchToProps = (dispatch: any) => ({
});

export default connect(mapStateToProps, mapDispatchToProps)(TransactionPage);
