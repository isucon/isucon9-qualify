import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { TransactionStatus } from '../../dataObjects/transaction';
import { ShippingStatus } from '../../dataObjects/shipping';
import Initial from './Buyer/Initial';
import WaitShipping from './Buyer/WaitShipping';
import WaitDone from './Buyer/WaitDone';
import Done from './Buyer/Done';

const useStyles = makeStyles(theme => ({
  progress: {
    top: 0,
    bottom: 0,
    right: 0,
    left: 0,
    margin: 'auto',
    position: 'absolute',
  },
}));

type Props = {
  itemId: number;
  postShipped: (itemId: number) => void;
  postComplete: (itemId: number) => void;
  transactionStatus: TransactionStatus;
  shippingStatus: ShippingStatus;
};

const BuyerComponent: React.FC<Props> = ({
  itemId,
  postShipped,
  postComplete,
  transactionStatus,
  shippingStatus,
}) => {
  const classes = useStyles();

  if (shippingStatus === 'initial' && transactionStatus === 'wait_shipping') {
    return <Initial itemId={itemId} postShipped={postShipped} />;
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

export default BuyerComponent;
