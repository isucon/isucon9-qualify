import React from 'react';
import { push } from 'connected-react-router';
import {SellingButonComponent} from "../components/SellingButtonComponent";
import {connect} from "react-redux";

const mapStateToProps = (state: any) => ({});

const mapDispatchToProps = (dispatch: any) => ({
    onClick: (e: React.MouseEvent) => {
        e.preventDefault();
        dispatch(push('/sell'));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(SellingButonComponent);