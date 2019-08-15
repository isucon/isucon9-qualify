import React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { TransactionStatus } from '../dataObjects/transaction';
import { ShippingStatus } from '../dataObjects/shipping';
import Paper from '@material-ui/core/Paper/Paper';

const useStyles = makeStyles(theme => ({
  soldOutLabel: {
    width: '30px',
    height: '30px',
    color: theme.palette.primary.light,
    backgroundColor: theme.palette.grey.A100,
  },
  tradingLabel: {
    width: '30px',
    height: '30px',
    color: theme.palette.primary.light,
    backgroundColor: theme.palette.primary.main,
  },
}));

type Props = {
  transactionStatus: TransactionStatus;
  shippingStatus: ShippingStatus;
};

const TransactionLabel: React.FC<Props> = ({
  transactionStatus,
  shippingStatus,
}: Props) => {
  const classes = useStyles();

  if (transactionStatus === 'done' && shippingStatus === 'done') {
    return <Paper className={classes.soldOutLabel}>売却済</Paper>;
  }

  return <Paper className={classes.tradingLabel}>取引中</Paper>;
};

export { TransactionLabel };
