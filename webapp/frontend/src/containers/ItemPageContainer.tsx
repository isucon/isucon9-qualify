import {connect} from "react-redux";
import ItemPage from "../pages/ItemPage";
import {fetchItemAction} from "../actions/fetchItemAction";
import {AppState} from "../index";

const mapStateToProps = (state: AppState) => ({
    isFetchingItem: state.viewingItem.isFetching,
});
const mapDispatchToProps = (dispatch: any) => ({
    load: (itemId: string) => {
        dispatch(fetchItemAction(itemId))
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(ItemPage);
