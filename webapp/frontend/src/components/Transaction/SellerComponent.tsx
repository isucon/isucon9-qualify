import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { TransactionStatus } from '../../dataObjects/transaction';
import { ShippingStatus } from '../../dataObjects/shipping';
import Initial from './Seller/Initial';
import WaitShipping from './Seller/WaitShipping';
import WaitDone from './Seller/WaitDone';
import Done from './Seller/Done';

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
  transactionEvidenceId: number;
  postShipped: (itemId: number) => void;
  transactionStatus: TransactionStatus;
  shippingStatus: ShippingStatus;
};

const SellerComponent: React.FC<Props> = ({
  itemId,
  transactionEvidenceId,
  postShipped,
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
    return (
      <WaitShipping
        itemId={itemId}
        transactionEvidenceId={transactionEvidenceId}
        postShipped={postShipped}
      />
    );
  }

  if (transactionStatus === 'wait_done') {
    return <WaitDone />;
  }

  return <Done />;
};

export default SellerComponent;
