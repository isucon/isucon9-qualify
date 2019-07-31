import React from 'react';
import { ItemListComponent } from "../components/ItemListComponent";
import { ItemData } from "../dataObjects/item";
import { UserData } from "../dataObjects/user";
import Avatar from "@material-ui/core/Avatar";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {Grid} from "@material-ui/core";
import Typography from "@material-ui/core/Typography";
import Divider from "@material-ui/core/Divider";

const useStyles = makeStyles(theme => ({
    avatar: {
        margin: theme.spacing(3),
        width: '100px',
        height: '100px',
    },
    itemList: {
        marginTop: theme.spacing(4),
    },
}));

interface UserPageProps {
    items: ItemData[]
    user: UserData
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

const mockUser: UserData = {
    id: 1235,
    accountName: 'Kirin',
    address: 'Tokyo',
};

const UserPage: React.FC/*<UserPageProps>*/ = (/*{ items, user }*/) => {
    const classes = useStyles();
    const user = mockUser;
    const items = mockItems;

    return (
        <div>
            <p>User Page</p>
            <Grid
                container
                direction="row"
                justify="center"
                alignItems="center"
                wrap="nowrap"
                spacing={2}
            >
                <Grid item>
                    <Avatar className={classes.avatar}>{user.accountName.charAt(0)}</Avatar>
                </Grid>
                <Grid item xs>
                    <Typography variant="h3">{user.accountName}</Typography>
                </Grid>
            </Grid>
            <Divider variant="middle" />
            <div className={classes.itemList}>
                <ItemListComponent items={items}/>
            </div>
        </div>
    );
};

export { UserPage }