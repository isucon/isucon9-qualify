import {connect} from "react-redux";
import ItemPage from "../pages/ItemPage";
import {fetchItemPageAction} from "../actions/fetchItemPageAction";
import {AppState} from "../index";
import {push} from "connected-react-router";
import {routes} from "../routes/Route";

const mapStateToProps = (state: AppState) => ({
    item: state.viewingItem.item,
    errorType: state.error.errorType,
    isLoading: state.page.isLoading,
});
const mapDispatchToProps = (dispatch: any) => ({
    load: (itemId: string) => {
        dispatch(fetchItemPageAction(itemId))
    },
    onClickBuy: (itemId: number) => {
        dispatch(push(routes.buy.getPath(itemId)));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(ItemPage);
