import React from 'react';
import {withBaseComponent} from "../hoc/withBaseComponent";
import ItemBuyFormContainer from "../containers/ItemBuyFormContainer";

const ItemBuyPage: React.FC = () => {
    return (
        <React.Fragment>
            <ItemBuyFormContainer />
        </React.Fragment>
    );
};

export default ItemBuyPage;