import React from 'react';
import { ItemListComponent } from "../components/ItemListComponent";
import { ItemData } from "../dataObjects/item";
import { UserData } from "../dataObjects/user";
import Avatar from "@material-ui/core/Avatar";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {Grid} from "@material-ui/core";
import Typography from "@material-ui/core/Typography";
import Divider from "@material-ui/core/Divider";
import SellingButtonContainer from "../containers/SellingButtonContainer";
import {BasePageComponent} from "../components/BasePageComponent";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";
import LoadingComponent from "../components/LoadingComponent";

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

type Props = {
    items: ItemData[]
    user: UserData,
    loading: boolean,
} & ErrorProps

const UserPage: React.FC<Props> = ({ items, user, loading }) => {
    const classes = useStyles();

    return (
        <BasePageComponent>
            {
                loading ?
                    (
                        <LoadingComponent />
                    ) : (
                        <React.Fragment>
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
                            <SellingButtonContainer />
                        </React.Fragment>
                )
            }
        </BasePageComponent>
    );
};

export default PageComponentWithError<Props>()(UserPage);