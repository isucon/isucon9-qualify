import React from 'react';
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import GridList from "@material-ui/core/GridList";
import { ItemComponent } from '../components/ItemComponent';
import GridListTile from "@material-ui/core/GridListTile";

const useStyles = makeStyles(theme => ({
    root: {
        display: 'flex',
        flexWrap: 'wrap',
        marginTop: theme.spacing(1),
        justifyContent: 'space-around',
        overflow: 'hidden',
    },
    grid: {
        width: '300px',
        height: '300px',
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

    const itemComponents = [];

    const items = mockItems;

    for (const item of items) {
        itemComponents.push(
            <GridListTile className={classes.grid} key={item.id}>
                <ItemComponent itemId={item.id} imageUrl={item.thumbnailUrl} title={item.name} price={item.price}/>
            </GridListTile>
        )
    }

    return (
        <div className={classes.root}>
            <GridList cols={3}>
                {itemComponents}
            </GridList>
        </div>
    );
};

export { ItemListPage }