import React from 'react';
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import { ItemListComponent } from '../components/ItemListComponent';
import SellingButtonContainer from "../containers/SellingButtonContainer";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";

const useStyles = makeStyles(theme => ({
    root: {
        display: 'flex',
        flexWrap: 'wrap',
        marginTop: theme.spacing(1),
        justifyContent: 'space-around',
        overflow: 'hidden',
    },
}));

type ItemListPageProps = {
    items: ItemData[],
} & ErrorProps

const mockItems: ItemData[] = [
    {
        id: 1,
        status: 'on_sale',
        sellerId: 1111,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
    {
        id: 2,
        status: 'on_sale',
        sellerId: 1111,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
    {
        id: 3,
        status: 'on_sale',
        sellerId: 1111,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
    {
        id: 4,
        status: 'on_sale',
        sellerId: 1111,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
];

const ItemListPage: React.FC/*<ItemListPageProps>*/ = (/*{ items }: ItemListPageProps*/) => {
    const classes = useStyles();
    const items = mockItems;

    return (
        <div className={classes.root}>
            <ItemListComponent items={items}/>
            <SellingButtonContainer />
        </div>
    );
};

export default PageComponentWithError<ItemListPageProps>()(ItemListPage);