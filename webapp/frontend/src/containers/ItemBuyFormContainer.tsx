import {connect} from "react-redux";
import ItemBuyFormComponent from "../components/ItemBuyFormComponent";
import {buyItemAction} from "../actions/buyAction";

const mapStateToProps = (state: any) => ({
    item: state.viewingItem.item,
    errors: state.formError.buyFormError,
});
const mapDispatchToProps = (dispatch: any) => ({
    onBuyAction: (itemId: number, cardNumber: string) => {
        dispatch(buyItemAction(itemId, cardNumber))
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(ItemBuyFormComponent);
