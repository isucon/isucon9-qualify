import React from 'react';
import { TransactionStatus } from '../../dataObjects/transaction';
import { ShippingStatus } from '../../dataObjects/shipping';
import Initial from '../Transaction/Seller/Initial';
import WaitShipping from '../Transaction/Seller/WaitShipping';
import WaitDone from '../Transaction/Seller/WaitDone';
import Done from '../Transaction/Seller/Done';

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
