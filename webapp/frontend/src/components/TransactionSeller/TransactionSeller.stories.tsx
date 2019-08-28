import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { TransactionSeller, Props } from '.';

const stories = storiesOf('components/TransactionSeller', module);

const mockProps: Props = {
  itemId: 1,
  transactionEvidenceId: 1,
  postShipped: (itemId: number) => {},
  postShippedDone: (itemId: number) => {},
  transactionStatus: 'wait_shipping',
  shippingStatus: 'initial',
};

stories
  .add('wait shipping', () => <TransactionSeller {...mockProps} />)
  .add('wait pickup', () => (
    <TransactionSeller {...mockProps} shippingStatus="wait_pickup" />
  ))
  .add('wait done', () => (
    <TransactionSeller {...mockProps} transactionStatus="wait_done" />
  ))
  .add('done', () => (
    <TransactionSeller {...mockProps} transactionStatus="done" />
  ));
