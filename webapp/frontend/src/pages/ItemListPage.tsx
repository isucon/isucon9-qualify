import React from 'react';
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import Card from "@material-ui/core/Card";
import GridList from "@material-ui/core/GridList";
import GridListTile from "@material-ui/core/GridListTile";
import GridListTileBar from "@material-ui/core/GridListTileBar";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    itemImage: {
        height: '100%',
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
            <GridListTile key={item.id}>
                <Card>
                    <img className={classes.itemImage} src={item.thumbnailUrl} alt={item.name} />
                    <GridListTileBar
                        title={item.name}
                        subtitle={`¥${item.price}`}
                    />
                </Card>
            </GridListTile>
        )
    }

    return (
        <div className={classes.paper}>
            <GridList cols={3}>
                {itemComponents}
            </GridList>
        </div>
    );
};

export { ItemListPage }