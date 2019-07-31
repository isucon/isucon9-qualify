import React from 'react';
import { ItemData } from "../dataObjects/item";
import {Typography} from "@material-ui/core";
import makeStyles from "@material-ui/core/styles/makeStyles";
import Grid from "@material-ui/core/Grid";
import Divider from "@material-ui/core/Divider";
import Avatar from "@material-ui/core/Avatar";
import { Link as RouteLink } from 'react-router-dom';
import AppBar from "@material-ui/core/AppBar";
import Button from "@material-ui/core/Button";

const useStyles = makeStyles(theme => ({
    title: {
        margin: theme.spacing(3),
    },
    itemImage: {
        width: '100%',
        maxWidth: '500px',
        height: 'auto',
    },
    avatar: {
        width: '50px',
        height: '50px',
    },
    divider: {
        margin: theme.spacing(1),
    },
    descSection: {
        marginTop: theme.spacing(3),
        marginBottom: theme.spacing(3),
    },
    link: {
        textDecoration: 'none',
    },
    appBar: {
        top: 'auto',
        bottom: 0,
    },
    buyButton: {
        margin: theme.spacing(1),
    }
}));

interface ItemPageProps {
    item: ItemData
}

const mockItem = {
    id: 1,
    name: 'いす',
    price: 10000,
    description: 'いすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすですいすです',
    createdAt: '2日前',
    thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
};

const ItemPage: React.FC/*<ItemPageProps>*/ = (/*{ item }*/) => {
    const classes = useStyles();

    const item = mockItem;

    return (
        <div>
            Item Page
            <Typography className={classes.title} variant="h3">{item.name}</Typography>
            <Grid container spacing={2}>
                <Grid item>
                    <img className={classes.itemImage} src={item.thumbnailUrl} />
                </Grid>
                <Grid item xs={12} sm container>
                    <Grid item xs container direction="column" spacing={2}>
                        <Grid item xs>
                            <div className={classes.descSection}>
                                <Typography variant="h4">商品説明</Typography>
                                <Divider className={classes.divider} variant="middle" />
                                <Typography variant="body1">{item.description}</Typography>
                            </div>

                            <div className={classes.descSection}>
                                <Typography variant="h4">出品者</Typography>
                                <Divider className={classes.divider} variant="middle" />
                                <Grid
                                    container
                                    direction="row"
                                    justify="center"
                                    alignItems="center"
                                    wrap="nowrap"
                                    spacing={2}
                                >
                                    <Grid item>
                                        <RouteLink className={classes.link} to={`/users/${1}`}>
                                            <Avatar className={classes.avatar}>"T"</Avatar>
                                        </RouteLink>
                                    </Grid>
                                    <Grid item xs>
                                        <Typography variant="body1">"TODO"</Typography>
                                    </Grid>
                                </Grid>
                            </div>
                        </Grid>
                    </Grid>
                </Grid>
            </Grid>
            <AppBar color="primary" position="fixed" className={classes.appBar}>
                <Grid
                    container
                    spacing={2}
                    direction="row"
                    alignItems="center"
                >
                    <Grid item>
                        <Typography variant="h5">¥{item.price}</Typography>
                    </Grid>
                    <Grid item>
                        <Button variant="contained" className={classes.buyButton}>購入</Button>
                    </Grid>
                </Grid>
            </AppBar>
        </div>
    );
};

export { ItemPage }