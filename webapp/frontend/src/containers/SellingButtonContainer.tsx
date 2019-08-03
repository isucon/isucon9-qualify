import { push } from 'connected-react-router';
import {SellingButonComponent} from "../components/SellingButtonComponent";
import {connect} from "react-redux";
import {routes} from "../routes/Route";

const mapStateToProps = (state: any) => ({});

const mapDispatchToProps = (dispatch: any) => ({
    onClick: (e: React.MouseEvent) => {
        e.preventDefault();
        dispatch(push(routes.sell.path));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SellingButonComponent);