import React from 'react';
import {ItemData} from "../dataObjects/item";

interface ItemListPageProps {
    items: ItemData[],
}

const ItemListPage: React.FC<ItemListPageProps> = ({ items }: ItemListPageProps) => {
    return (
        <div>Item list Page</div>
    );
};

export { ItemListPage }