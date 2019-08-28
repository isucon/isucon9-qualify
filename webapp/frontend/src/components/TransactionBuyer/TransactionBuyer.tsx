import React from 'react';
import {makeStyles, Theme} from '@material-ui/core/styles';
import { TransactionStatus } from '../../dataObjects/transaction';
import { ShippingStatus } from '../../dataObjects/shipping';
import Initial from '../Transaction/Buyer/Initial';
import WaitShipping from '../Transaction/Buyer/WaitShipping';
import WaitDone from '../Transaction/Buyer/WaitDone';
import Done from '../Transaction/Buyer/Done';

const useStyles = makeStyles((theme: Theme) => ({
  progress: {
    top: 0,
    bottom: 0,
    right: 0,
    left: 0,
    margin: 'auto',
    position: 'absolute',
  },
}));

export type Props = {
  itemId: number;
  postComplete: (itemId: number) => void;
  transactionStatus: TransactionStatus;
  shippingStatus: ShippingStatus;
};

const TransactionBuyer: React.FC<Props> = ({
  itemId,
  postComplete,
  transactionStatus,
  shippingStatus,
}) => {
  const classes = useStyles();

  if (shippingStatus === 'initial' && transactionStatus === 'wait_shipping') {
    return <Initial />;
  }

  if (
    shippingStatus === 'wait_pickup' &&
    transactionStatus === 'wait_shipping'
  ) {
    return <WaitShipping />;
  }

  if (transactionStatus === 'wait_done') {
    return <WaitDone itemId={itemId} postComplete={postComplete} />;
  }

  return <Done />;
};

export { TransactionBuyer };
