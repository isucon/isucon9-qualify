import React from 'react';
import { TransactionStatus } from '../../dataObjects/transaction';
import { ShippingStatus } from '../../dataObjects/shipping';
import Initial from '../Transaction/Buyer/Initial';
import WaitShipping from '../Transaction/Buyer/WaitShipping';
import WaitDone from '../Transaction/Buyer/WaitDone';
import Done from '../Transaction/Buyer/Done';

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
