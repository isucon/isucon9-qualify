import React from 'react';
import {withBaseComponent} from "../hoc/withBaseComponent";
import {ItemData} from "../dataObjects/item";
import {BuyFormErrorState} from "../reducers/formErrorReducer";
import ItemBuyFormComponent from "../components/ItemBuyFormComponent";

interface ItemBuyPageProps {
    item: ItemData,
}

const ItemBuyPage: React.FC/*<ItemBuyPageProps>*/ = (/*{ item }*/) => {
    const errors = {
        cardError: [],
        buyError: [],
    };
    const item = {
        id: 1,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    };

    return (
        <React.Fragment>
            <ItemBuyFormComponent item={item} errors={errors}/>
        </React.Fragment>
    );
};

export default withBaseComponent(ItemBuyPage);