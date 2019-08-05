import {connect} from "react-redux";
import ItemPage from "../pages/ItemPage";
import {fetchItemPageAction} from "../actions/fetchItemPageAction";
import {AppState} from "../index";

const mapStateToProps = (state: AppState) => ({
    errorType: state.error.errorType,
    isLoading: state.page.isLoading,
});
const mapDispatchToProps = (dispatch: any) => ({
    load: (itemId: string) => {
        dispatch(fetchItemPageAction(itemId))
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(ItemPage);
