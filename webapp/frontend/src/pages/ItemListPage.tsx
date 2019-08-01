import React from 'react';
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import { ItemListComponent } from '../components/ItemListComponent';

const useStyles = makeStyles(theme => ({
    root: {
        display: 'flex',
        flexWrap: 'wrap',
        marginTop: theme.spacing(1),
        justifyContent: 'space-around',
        overflow: 'hidden',
    },
}));

interface ItemListPageProps {
    items: ItemData[],
}

const mockItems: ItemData[] = [
    {
        id: 1,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
    {
        id: 2,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
    {
        id: 3,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    },
    {
        id: 4,
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
        </div>
    );
};

export { ItemListPage }