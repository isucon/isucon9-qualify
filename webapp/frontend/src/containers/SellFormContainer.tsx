import {connect} from "react-redux";
import SellFormComponent from "../components/SellFormComponent";
import {listItemAction} from "../actions/sellingItemAction";
import {AppState} from "../index";
import {AnyAction} from "redux";
import {ThunkDispatch} from "redux-thunk";

const mapStateToProps = (state: AppState) => ({
    error: state.formError.error,
    categories: [ // TODO
        {
            id: 1,
            categoryName: 'カテゴリ1',
        },
        {
            id: 2,
            categoryName: 'カテゴリ2',
        }
    ],
});
const mapDispatchToProps = (dispatch: ThunkDispatch<AppState, undefined, AnyAction>) => ({
    sellItem: (name: string, description: string, price: number, categoryId: number) => {
        dispatch(listItemAction(name, description, price, categoryId));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SellFormComponent);
