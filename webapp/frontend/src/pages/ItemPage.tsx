import React from 'react';
import { ItemData } from "../dataObjects/item";
import {createStyles, Theme, Typography, WithStyles} from "@material-ui/core";
import Grid from "@material-ui/core/Grid";
import Divider from "@material-ui/core/Divider";
import Avatar from "@material-ui/core/Avatar";
import {Link, Link as RouteLink, RouteComponentProps} from 'react-router-dom';
import AppBar from "@material-ui/core/AppBar";
import Button from "@material-ui/core/Button";
import {routes} from "../routes/Route";
import {StyleRules} from "@material-ui/core/styles";
import withStyles from "@material-ui/core/styles/withStyles";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";
import BasePageContainer from "../containers/BasePageContainer";
import LoadingComponent from "../components/LoadingComponent";

const styles = (theme: Theme): StyleRules => createStyles({
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
});

interface ItemPageProps extends WithStyles<typeof styles> {
    loading: boolean,
    item: ItemData
    load: (itemId: string) => void
    onClickBuy: (itemId: number) => void
}

type Props = ItemPageProps & RouteComponentProps<{ item_id: string }> & ErrorProps

class ItemPage extends React.Component<Props> {
    constructor(props: Props) {
        super(props);

        this.props.load(this.props.match.params.item_id);
        this._onClickBuyButton = this._onClickBuyButton.bind(this);
    }

    _onClickBuyButton(e: React.MouseEvent) {
        e.preventDefault();
        this.props.onClickBuy(this.props.item.id);
    }

    render() {
        const { classes, item, loading } = this.props;

        return (
            <BasePageContainer>
                {
                    loading ? (
                        <LoadingComponent/>
                    ) : (
                        <React.Fragment>
                            Item Page
                            <Typography className={classes.title} variant="h3">{item.name}</Typography>
                            <Grid container spacing={2}>
                                <Grid item>
                                    <img className={classes.itemImage} alt={item.name} src={item.thumbnailUrl}/>
                                </Grid>
                                <Grid item xs={12} sm container>
                                    <Grid item xs container direction="column" spacing={2}>
                                        <Grid item xs>
                                            <div className={classes.descSection}>
                                                <Typography variant="h4">商品説明</Typography>
                                                <Divider className={classes.divider} variant="middle"/>
                                                <Typography variant="body1">{item.description}</Typography>
                                            </div>

                                            <div className={classes.descSection}>
                                                <Typography variant="h4">カテゴリ</Typography>
                                                <Divider className={classes.divider} variant="middle"/>
                                                <Typography variant="body1">
                                                    <Link to={routes.categoryTimeline.getPath(item.category.parentId)}>
                                                        {item.category.parentCategoryName}
                                                    </Link> > {item.category.categoryName}
                                                </Typography>
                                            </div>

                                            <div className={classes.descSection}>
                                                <Typography variant="h4">出品者</Typography>
                                                <Divider className={classes.divider} variant="middle"/>
                                                <Grid
                                                    container
                                                    direction="row"
                                                    justify="center"
                                                    alignItems="center"
                                                    wrap="nowrap"
                                                    spacing={2}
                                                >
                                                    <Grid item>
                                                        <RouteLink className={classes.link}
                                                                   to={routes.user.getPath(item.sellerId)}>
                                                            <Avatar
                                                                className={classes.avatar}>{item.seller.accountName.charAt(0)}</Avatar>
                                                        </RouteLink>
                                                    </Grid>
                                                    <Grid item xs>
                                                        <Typography variant="body1">{item.seller.accountName}</Typography>
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
                                        <Button
                                            variant="contained"
                                            className={classes.buyButton}
                                            onClick={this._onClickBuyButton}
                                        >
                                            購入
                                        </Button>
                                    </Grid>
                                </Grid>
                            </AppBar>
                        </React.Fragment>
                    )
                }
            </BasePageContainer>
        );
    }
}

export default PageComponentWithError<any>()(withStyles(styles)(ItemPage));
