import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { TransactionBuyer, Props } from '.';

const stories = storiesOf('components/TransactionBuyer', module);

const mockProps: Props = {
  itemId: 1,
  postComplete: (itemId: number) => {},
  transactionStatus: 'wait_shipping',
  shippingStatus: 'initial',
};

stories
  .add('wait shipping', () => <TransactionBuyer {...mockProps} />)
  .add('wait pickup', () => (
    <TransactionBuyer {...mockProps} shippingStatus="wait_pickup" />
  ))
  .add('wait done', () => (
    <TransactionBuyer {...mockProps} transactionStatus="wait_done" />
  ))
  .add('done', () => (
    <TransactionBuyer {...mockProps} transactionStatus="done" />
  ));
