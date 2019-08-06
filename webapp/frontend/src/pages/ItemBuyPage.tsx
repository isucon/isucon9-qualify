import React from 'react';
import ItemBuyFormContainer from "../containers/ItemBuyFormContainer";
import {BasePageComponent} from "../components/BasePageComponent";

const ItemBuyPage: React.FC = () => {
    return (
        <BasePageComponent>
            <ItemBuyFormContainer />
        </BasePageComponent>
    );
};

export default ItemBuyPage;