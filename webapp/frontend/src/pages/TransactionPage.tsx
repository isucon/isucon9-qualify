import React, { ReactElement } from 'react';
import BasePageContainer from '../containers/BasePageContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import { RouteComponentProps } from 'react-router';
import { ItemData } from '../dataObjects/item';
import LoadingComponent from '../components/LoadingComponent';
import { NotFoundPage } from './error/NotFoundPage';
import SellerTransactionContainer from '../containers/SellerTransactionContainer';
import { InternalServerErrorPage } from './error/InternalServerErrorPage';
import BuyerTransactionContainer from '../containers/BuyerTransactionContainer';
import { createStyles, Grid, Theme, WithStyles } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import { routes } from '../routes/Route';
import { Link as RouteLink } from 'react-router-dom';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    itemImage: {
      width: '100%',
      maxWidth: '500px',
      height: 'auto',
    },
    divider: {
      margin: theme.spacing(1),
    },
    descSection: {
      marginTop: theme.spacing(3),
      marginBottom: theme.spacing(3),
    },
  });

interface TransactionProps extends WithStyles<typeof styles> {
  loading: boolean;
  item?: ItemData;
  load: (itemId: string) => void;
  // Logged in user info
  auth: {
    userId: number;
  };
}

type Props = TransactionProps &
  RouteComponentProps<{ item_id: string }> &
  ErrorProps;

class TransactionPage extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    this.props.load(this.props.match.params.item_id);
  }

  render() {
    const {
      loading,
      item,
      auth: { userId },
      classes,
    } = this.props;

    if (loading) {
      return (
        <BasePageContainer>
          <LoadingComponent />
        </BasePageContainer>
      );
    }

    if (item === undefined) {
      return <NotFoundPage />;
    }

    if (
      item.shippingStatus === undefined ||
      item.transactionEvidenceStatus === undefined ||
      item.transactionEvidenceId === undefined
    ) {
      return (
        <InternalServerErrorPage message="取引中の商品ではない、もしくはデータ形式が不正です" />
      );
    }

    let TransactionComponent: ReactElement | undefined;

    if (userId === item.sellerId) {
      TransactionComponent = (
        <SellerTransactionContainer
          itemId={item.id}
          transactionEvidenceId={item.transactionEvidenceId}
          transactionStatus={item.transactionEvidenceStatus}
          shippingStatus={item.shippingStatus}
        />
      );
    }

    if (userId === item.buyerId) {
      TransactionComponent = (
        <BuyerTransactionContainer
          itemId={item.id}
          transactionStatus={item.transactionEvidenceStatus}
          shippingStatus={item.shippingStatus}
        />
      );
    }

    if (TransactionComponent === undefined) {
      return <NotFoundPage message="商品が読み込めませんでした" />;
    }

    return (
      <BasePageContainer>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            {TransactionComponent}
          </Grid>
          <Grid item xs={12}>
            <Typography className={classes.descSection} variant="h4">
              取引情報
            </Typography>
            <Divider className={classes.divider} variant="middle" />
          </Grid>
          <Grid item xs={3}>
            <RouteLink to={routes.item.getPath(item.id)}>
              <img
                className={classes.itemImage}
                alt={item.name}
                src={item.thumbnailUrl}
              />
            </RouteLink>
          </Grid>
          <Grid item xs={9} container>
            <Grid item>
              <Typography variant="h5">{item.name}</Typography>
              <Typography variant="h6">{item.price}ｲｽｺｲﾝ</Typography>
            </Grid>
          </Grid>
        </Grid>
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(
  withStyles(styles)(TransactionPage),
);
