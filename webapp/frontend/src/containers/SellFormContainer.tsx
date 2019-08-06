import {connect} from "react-redux";
import SellFormComponent from "../components/SellFormComponent";
import {listItemAction} from "../actions/sellingItemAction";

const mapStateToProps = (state: any) => ({
    error: state.formError.error,
});
const mapDispatchToProps = (dispatch: any) => ({
    sellItem: (name: string, description: string, price: number) => {
        dispatch(listItemAction(name, description, price));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SellFormComponent);
