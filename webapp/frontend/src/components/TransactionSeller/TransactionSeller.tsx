import React from 'react';
import {makeStyles, Theme} from '@material-ui/core/styles';
import { TransactionStatus } from '../../dataObjects/transaction';
import { ShippingStatus } from '../../dataObjects/shipping';
import Initial from '../Transaction/Seller/Initial';
import WaitShipping from '../Transaction/Seller/WaitShipping';
import WaitDone from '../Transaction/Seller/WaitDone';
import Done from '../Transaction/Seller/Done';

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
  transactionEvidenceId: number;
  postShipped: (itemId: number) => void;
  postShippedDone: (itemId: number) => void;
  transactionStatus: TransactionStatus;
  shippingStatus: ShippingStatus;
};

const TransactionSeller: React.FC<Props> = ({
  itemId,
  transactionEvidenceId,
  postShipped,
  postShippedDone,
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
    return (
      <WaitShipping
        itemId={itemId}
        transactionEvidenceId={transactionEvidenceId}
        postShippedDone={postShippedDone}
      />
    );
  }

  if (transactionStatus === 'wait_done') {
    return <WaitDone />;
  }

  return <Done />;
};

export { TransactionSeller };
