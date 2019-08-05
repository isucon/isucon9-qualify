import {connect} from "react-redux";
import {BasePageComponent} from "../components/BasePageComponent";
import {AppState} from "../index";

const mapStateToProps = (state: AppState) => ({
    errorType: state.error.errorType,
    isLoading: state.viewingItem.isFetching, // TODO add new reducer for this value
});

const mapDispatchToProps = (dispatch: any) => ({});

export default connect(mapStateToProps, mapDispatchToProps)(BasePageComponent);

