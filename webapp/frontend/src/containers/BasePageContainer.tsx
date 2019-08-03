import {connect} from "react-redux";
import {BasePageComponent} from "../components/BasePageComponent";
import {AppState} from "../index";

const mapStateToProps = (state: AppState) => ({
    errorType: state.error.errorType,
});
const mapDispatchToProps = (dispatch: any) => ({});

export default connect(mapStateToProps, mapDispatchToProps)(BasePageComponent);

